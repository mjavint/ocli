package db

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/lib/pq"
	"github.com/sirupsen/logrus"
)

var log = logrus.New()

// PGConfig represents PostgreSQL connection configuration
type PGConfig struct {
	Host            string
	Port            int
	User            string
	Password        string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

// DefaultPGConfig returns a config with sensible defaults
func DefaultPGConfig() *PGConfig {
	return &PGConfig{
		Host:            "localhost",
		Port:            5432,
		User:            "postgres",
		SSLMode:         "disable",
		MaxOpenConns:    25,
		MaxIdleConns:    5,
		ConnMaxLifetime: 5 * time.Minute,
		ConnMaxIdleTime: 1 * time.Minute,
	}
}

// Connect connects to a PostgreSQL database with context support
func Connect(ctx context.Context, dbname string, cfg *PGConfig) (*sql.DB, error) {
	if cfg == nil {
		return nil, fmt.Errorf("postgres config is nil")
	}

	sslMode := cfg.SSLMode
	if sslMode == "" {
		sslMode = "disable"
	}

	connStr := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host,
		cfg.Port,
		cfg.User,
		cfg.Password,
		dbname,
		sslMode,
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open connection: %w", err)
	}

	// Configure connection pool
	if cfg.MaxOpenConns > 0 {
		db.SetMaxOpenConns(cfg.MaxOpenConns)
	}
	if cfg.MaxIdleConns > 0 {
		db.SetMaxIdleConns(cfg.MaxIdleConns)
	}
	if cfg.ConnMaxLifetime > 0 {
		db.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	}
	if cfg.ConnMaxIdleTime > 0 {
		db.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)
	}

	// Verify connection with context
	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}

// CloseDB closes a database connection safely
func CloseDB(db *sql.DB) error {
	if db != nil {
		return db.Close()
	}
	return nil
}

// DBExists checks if a database exists
func DBExists(ctx context.Context, dbname string, cfg *PGConfig) (bool, error) {
	db, err := Connect(ctx, "postgres", cfg)
	if err != nil {
		return false, fmt.Errorf("failed to connect to postgres: %w", err)
	}
	defer CloseDB(db)

	var exists bool
	query := "SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = $1)"
	err = db.QueryRowContext(ctx, query, dbname).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check database existence: %w", err)
	}

	return exists, nil
}

// GetInstalledModules returns names of all installed modules
func GetInstalledModules(ctx context.Context, db *sql.DB) ([]string, error) {
	query := `
		SELECT name FROM ir_module_module
		WHERE state IN ('installed', 'to upgrade')
		ORDER BY name
	`

	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query modules: %w", err)
	}
	defer rows.Close()

	var modules []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, fmt.Errorf("failed to scan module name: %w", err)
		}
		modules = append(modules, name)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating modules: %w", err)
	}

	return modules, nil
}

// CreateDatabase creates a new database
func CreateDatabase(ctx context.Context, dbname string, cfg *PGConfig) error {
	if !IsValidDBName(dbname) {
		return fmt.Errorf("invalid database name: %s", dbname)
	}

	db, err := Connect(ctx, "postgres", cfg)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer CloseDB(db)

	query := fmt.Sprintf("CREATE DATABASE %s ENCODING 'UTF8'",
		pq.QuoteIdentifier(dbname))

	_, err = db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to create database: %w", err)
	}

	log.WithField("database", dbname).Info("Database created successfully")
	return nil
}

// DropDatabase drops a database
func DropDatabase(ctx context.Context, dbname string, cfg *PGConfig) error {
	db, err := Connect(ctx, "postgres", cfg)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer CloseDB(db)

	// Terminate active connections
	if err := terminateConnections(ctx, db, dbname); err != nil {
		log.WithError(err).Warn("Failed to terminate connections, continuing anyway")
	}

	query := fmt.Sprintf("DROP DATABASE IF EXISTS %s",
		pq.QuoteIdentifier(dbname))

	_, err = db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to drop database: %w", err)
	}

	log.WithField("database", dbname).Info("Database dropped successfully")
	return nil
}

// terminateConnections terminates all connections to a database
func terminateConnections(ctx context.Context, db *sql.DB, dbname string) error {
	query := `
		SELECT pg_terminate_backend(pid)
		FROM pg_stat_activity
		WHERE datname = $1 AND pid <> pg_backend_pid()
	`
	_, err := db.ExecContext(ctx, query, dbname)
	return err
}

// CreateDatabaseFromTemplate creates a database from a template
func CreateDatabaseFromTemplate(ctx context.Context, dbname, template string, cfg *PGConfig) error {
	if !IsValidDBName(dbname) {
		return fmt.Errorf("invalid database name: %s", dbname)
	}

	db, err := Connect(ctx, "postgres", cfg)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer CloseDB(db)

	if err := terminateConnections(ctx, db, template); err != nil {
		log.WithError(err).Warn("Failed to terminate template connections")
	}

	query := fmt.Sprintf("CREATE DATABASE %s WITH TEMPLATE %s",
		pq.QuoteIdentifier(dbname),
		pq.QuoteIdentifier(template))

	_, err = db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to create database from template: %w", err)
	}

	log.WithFields(logrus.Fields{
		"database": dbname,
		"template": template,
	}).Info("Database created from template successfully")

	return nil
}

// RenameDatabase renames a database
func RenameDatabase(ctx context.Context, oldName, newName string, cfg *PGConfig) error {
	if !IsValidDBName(newName) {
		return fmt.Errorf("invalid new database name: %s", newName)
	}

	db, err := Connect(ctx, "postgres", cfg)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer CloseDB(db)

	if err := terminateConnections(ctx, db, oldName); err != nil {
		log.WithError(err).Warn("Failed to terminate connections")
	}

	query := fmt.Sprintf("ALTER DATABASE %s RENAME TO %s",
		pq.QuoteIdentifier(oldName),
		pq.QuoteIdentifier(newName))

	_, err = db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to rename database: %w", err)
	}

	log.WithFields(logrus.Fields{
		"old_name": oldName,
		"new_name": newName,
	}).Info("Database renamed successfully")

	return nil
}

// CopyDatabase copies a database
func CopyDatabase(ctx context.Context, source, target string, cfg *PGConfig) error {
	return CreateDatabaseFromTemplate(ctx, target, source, cfg)
}

// PingPostgres tests PostgreSQL connection
func PingPostgres(ctx context.Context, cfg *PGConfig) error {
	db, err := Connect(ctx, "postgres", cfg)
	if err != nil {
		return err
	}
	defer CloseDB(db)

	return db.PingContext(ctx)
}

// IsInitialized checks if a database is initialized (has Odoo tables)
func IsInitialized(ctx context.Context, dbname string, cfg *PGConfig) (bool, error) {
	exists, err := DBExists(ctx, dbname, cfg)
	if err != nil {
		return false, err
	}
	if !exists {
		return false, nil
	}

	db, err := Connect(ctx, dbname, cfg)
	if err != nil {
		return false, fmt.Errorf("failed to connect to database: %w", err)
	}
	defer CloseDB(db)

	var tableExists bool
	query := `
		SELECT EXISTS (
			SELECT 1 FROM information_schema.tables
			WHERE table_schema = 'public'
			AND table_name = 'ir_module_module'
		)
	`
	err = db.QueryRowContext(ctx, query).Scan(&tableExists)
	if err != nil {
		return false, fmt.Errorf("failed to check table existence: %w", err)
	}

	return tableExists, nil
}

// ResetConfigParameters resets database config parameters for neutralization
func ResetConfigParameters(ctx context.Context, db *sql.DB) error {
	queries := []string{
		"UPDATE ir_cron SET active = false",
		"UPDATE ir_mail_server SET active = false",
		"UPDATE ir_config_parameter SET value = 'test' WHERE key LIKE '%api_key%'",
		"UPDATE ir_config_parameter SET value = 'http://localhost' WHERE key LIKE '%base_url%'",
	}

	for _, query := range queries {
		_, err := db.ExecContext(ctx, query)
		if err != nil {
			log.WithError(err).WithField("query", query).Warn("Failed to execute neutralization query")
		}
	}

	log.Info("Database neutralized successfully")
	return nil
}

// ListDatabases lists all databases
func ListDatabases(ctx context.Context, cfg *PGConfig) ([]string, error) {
	db, err := Connect(ctx, "postgres", cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to connect: %w", err)
	}
	defer CloseDB(db)

	query := `
		SELECT datname FROM pg_database
		WHERE datistemplate = false
		AND datname != 'postgres'
		ORDER BY datname
	`

	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query databases: %w", err)
	}
	defer rows.Close()

	var databases []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, fmt.Errorf("failed to scan database name: %w", err)
		}
		databases = append(databases, name)
	}

	return databases, rows.Err()
}

// ListInitializedDatabases lists only Odoo-initialized databases
func ListInitializedDatabases(ctx context.Context, cfg *PGConfig) ([]string, error) {
	allDBs, err := ListDatabases(ctx, cfg)
	if err != nil {
		return nil, err
	}

	var initialized []string
	for _, dbname := range allDBs {
		if isInit, err := IsInitialized(ctx, dbname, cfg); err == nil && isInit {
			initialized = append(initialized, dbname)
		}
	}

	return initialized, nil
}

// GetDatabaseSize returns the size of a database in a human-readable format
func GetDatabaseSize(ctx context.Context, dbname string, cfg *PGConfig) (string, error) {
	db, err := Connect(ctx, "postgres", cfg)
	if err != nil {
		return "", err
	}
	defer CloseDB(db)

	var size int64
	query := "SELECT pg_database_size($1)"
	err = db.QueryRowContext(ctx, query, dbname).Scan(&size)
	if err != nil {
		return "", err
	}

	return formatBytes(size), nil
}

// formatBytes converts bytes to a human-readable format
func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}

	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	units := []string{"KB", "MB", "GB", "TB", "PB"}
	return fmt.Sprintf("%.2f %s", float64(bytes)/float64(div), units[exp])
}

// SetConfigParameter sets an ir_config_parameter value
func SetConfigParameter(ctx context.Context, db *sql.DB, key, value string) error {
	query := `
		INSERT INTO ir_config_parameter (key, value, create_uid, create_date, write_uid, write_date)
		VALUES ($1, $2, 1, NOW(), 1, NOW())
		ON CONFLICT (key)
		DO UPDATE SET value = $2, write_date = NOW(), write_uid = 1
	`

	_, err := db.ExecContext(ctx, query, key, value)
	if err != nil {
		return fmt.Errorf("failed to set config parameter: %w", err)
	}

	return nil
}

// GetConfigParameter gets an ir_config_parameter value
func GetConfigParameter(ctx context.Context, db *sql.DB, key string) (string, error) {
	var value string
	query := "SELECT value FROM ir_config_parameter WHERE key = $1"
	err := db.QueryRowContext(ctx, query, key).Scan(&value)

	if err == sql.ErrNoRows {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("failed to get config parameter: %w", err)
	}

	return value, nil
}

// IsValidDBName validates a PostgreSQL database name
func IsValidDBName(name string) bool {
	if len(name) == 0 || len(name) > 63 {
		return false
	}

	if !isAlphaOrUnderscore(rune(name[0])) {
		return false
	}

	for _, c := range name {
		if !isAlphanumericOrUnderscore(c) {
			return false
		}
	}

	reserved := []string{"template0", "template1", "postgres"}
	nameLower := strings.ToLower(name)
	for _, r := range reserved {
		if nameLower == r {
			return false
		}
	}

	return true
}

func isAlphaOrUnderscore(c rune) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || c == '_'
}

func isAlphanumericOrUnderscore(c rune) bool {
	return isAlphaOrUnderscore(c) || (c >= '0' && c <= '9')
}
