package db

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/lib/pq"
	"github.com/mjavint/ocli/pkg/constants"
	apperrors "github.com/mjavint/ocli/pkg/errors"
	"github.com/mjavint/ocli/pkg/logger"
	"github.com/sirupsen/logrus"
)

// Querier is the minimal interface required to query database sizes.
type Querier interface {
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
}

// DatabaseSize holds a database name and its formatted size.
type DatabaseSize struct {
	Name string
	Size string
}

// Global logger instance for backward compatibility
var log = logger.NewLogger().GetLogger()

// PostgresDB implements database operations for PostgreSQL
type PostgresDB struct {
	config *PGConfig
	logger *logger.Logger
}

// NewPostgresDB creates a new PostgreSQL database manager
func NewPostgresDB(config *PGConfig, log *logger.Logger) *PostgresDB {
	if log == nil {
		log = logger.NewLogger()
	}
	return &PostgresDB{
		config: config,
		logger: log,
	}
}

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

// GetHost returns the database host
func (c *PGConfig) GetHost() string {
	return c.Host
}

// GetPort returns the database port
func (c *PGConfig) GetPort() int {
	return c.Port
}

// GetUser returns the database user
func (c *PGConfig) GetUser() string {
	return c.User
}

// GetPassword returns the database password
func (c *PGConfig) GetPassword() string {
	return c.Password
}

// GetSSLMode returns the SSL mode
func (c *PGConfig) GetSSLMode() string {
	return c.SSLMode
}

// GetMaxOpenConns returns the maximum open connections
func (c *PGConfig) GetMaxOpenConns() int {
	return c.MaxOpenConns
}

// GetMaxIdleConns returns the maximum idle connections
func (c *PGConfig) GetMaxIdleConns() int {
	return c.MaxIdleConns
}

// GetConnMaxLifetime returns the connection maximum lifetime
func (c *PGConfig) GetConnMaxLifetime() time.Duration {
	return c.ConnMaxLifetime
}

// GetConnMaxIdleTime returns the connection maximum idle time
func (c *PGConfig) GetConnMaxIdleTime() time.Duration {
	return c.ConnMaxIdleTime
}

// DefaultPGConfig returns a config with sensible defaults
func DefaultPGConfig() *PGConfig {
	return &PGConfig{
		Host:            constants.DefaultDBHost,
		Port:            constants.DefaultDBPort,
		User:            constants.DefaultDBUser,
		SSLMode:         constants.DefaultSSLMode,
		MaxOpenConns:    constants.DefaultMaxOpenConns,
		MaxIdleConns:    constants.DefaultMaxIdleConns,
		ConnMaxLifetime: constants.DefaultConnMaxLifetime,
		ConnMaxIdleTime: constants.DefaultConnMaxIdleTime,
	}
}

// Connect connects to a PostgreSQL database with context support
func (pg *PostgresDB) Connect(ctx context.Context, dbname string) (*sql.DB, error) {
	if pg.config == nil {
		return nil, apperrors.NewInvalidConfigError("postgres config is nil", nil)
	}

	sslMode := pg.config.SSLMode
	if sslMode == "" {
		sslMode = constants.DefaultSSLMode
	}

	connStr := pg.buildConnectionString(dbname, sslMode)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, apperrors.NewDatabaseConnectionError("failed to open connection", err)
	}

	pg.configureConnectionPool(db)

	// Verify connection with context
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, apperrors.NewDatabaseConnectionError("failed to ping database", err)
	}

	return db, nil
}

// buildConnectionString builds the PostgreSQL connection string
func (pg *PostgresDB) buildConnectionString(dbname, sslMode string) string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		pg.config.Host,
		pg.config.Port,
		pg.config.User,
		pg.config.Password,
		dbname,
		sslMode,
	)
}

// configureConnectionPool configures the database connection pool
func (pg *PostgresDB) configureConnectionPool(db *sql.DB) {
	if pg.config.MaxOpenConns > 0 {
		db.SetMaxOpenConns(pg.config.MaxOpenConns)
	}
	if pg.config.MaxIdleConns > 0 {
		db.SetMaxIdleConns(pg.config.MaxIdleConns)
	}
	if pg.config.ConnMaxLifetime > 0 {
		db.SetConnMaxLifetime(pg.config.ConnMaxLifetime)
	}
	if pg.config.ConnMaxIdleTime > 0 {
		db.SetConnMaxIdleTime(pg.config.ConnMaxIdleTime)
	}
}

// CloseDB closes a database connection safely
func CloseDB(db *sql.DB) error {
	if db != nil {
		return db.Close()
	}
	return nil
}

// Close closes a database connection (method version)
func (pg *PostgresDB) Close(db *sql.DB) error {
	return CloseDB(db)
}

// Connect is a legacy wrapper for backward compatibility
func Connect(ctx context.Context, dbname string, cfg *PGConfig) (*sql.DB, error) {
	pg := NewPostgresDB(cfg, nil)
	return pg.Connect(ctx, dbname)
}

// DBExists checks if a database exists
func DBExists(ctx context.Context, dbname string, cfg *PGConfig) (bool, error) {
	db, err := Connect(ctx, "postgres", cfg)
	if err != nil {
		return false, fmt.Errorf("failed to connect to postgres: %w", err)
	}
	defer func() { _ = CloseDB(db) }()

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
	defer func() { _ = rows.Close() }()

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
	defer func() { _ = CloseDB(db) }()

	query := fmt.Sprintf("CREATE DATABASE %s ENCODING 'UTF8",
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
	defer func() { _ = CloseDB(db) }()

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
	defer func() { _ = CloseDB(db) }()

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
	defer func() { _ = CloseDB(db) }()

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
	defer func() { _ = CloseDB(db) }()

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
	defer func() { _ = CloseDB(db) }()

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
	defer func() { _ = CloseDB(db) }()

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
	defer func() { _ = rows.Close() }()

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
	defer func() { _ = CloseDB(db) }()

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
	if len(name) == 0 || len(name) > constants.MaxDBNameLength {
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

// ListDatabasesWithSizes returns non-template user databases with sizes in a single query.
func ListDatabasesWithSizes(ctx context.Context, q Querier) ([]DatabaseSize, error) {
	query := `
		SELECT datname, pg_database_size(datname)
		FROM pg_database
		WHERE datistemplate = false
		  AND datname != 'postgres'
		ORDER BY datname`
	rows, err := q.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query databases: %w", err)
	}
	if rows == nil {
		return []DatabaseSize{}, nil
	}
	defer func() { _ = rows.Close() }()

	var out []DatabaseSize
	for rows.Next() {
		var name string
		var sizeBytes int64
		if err := rows.Scan(&name, &sizeBytes); err != nil {
			return nil, fmt.Errorf("failed to scan database: %w", err)
		}
		out = append(out, DatabaseSize{Name: name, Size: formatBytes(sizeBytes)})
	}
	return out, rows.Err()
}

// ListDatabasesWithSizesConfig connects once and returns databases with sizes.
func ListDatabasesWithSizesConfig(ctx context.Context, cfg *PGConfig) ([]DatabaseSize, error) {
	dbConn, err := Connect(ctx, "postgres", cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to connect: %w", err)
	}
	defer func() { _ = CloseDB(dbConn) }()
	return ListDatabasesWithSizes(ctx, dbConn)
}
