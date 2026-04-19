package interfaces

import (
	"context"
	"database/sql"
	"os"
	"time"
)

// DatabaseConfig defines the interface for database configuration
type DatabaseConfig interface {
	GetHost() string
	GetPort() int
	GetUser() string
	GetPassword() string
	GetSSLMode() string
	GetMaxOpenConns() int
	GetMaxIdleConns() int
	GetConnMaxLifetime() time.Duration
	GetConnMaxIdleTime() time.Duration
}

// DatabaseConnection defines the interface for database operations
type DatabaseConnection interface {
	Connect(ctx context.Context, dbname string) (*sql.DB, error)
	Close(db *sql.DB) error
	Ping(ctx context.Context, db *sql.DB) error
}

// DatabaseManager defines the interface for database management operations
type DatabaseManager interface {
	Exists(ctx context.Context, dbname string) (bool, error)
	Create(ctx context.Context, dbname string) error
	Drop(ctx context.Context, dbname string) error
	Rename(ctx context.Context, oldName, newName string) error
	Copy(ctx context.Context, source, target string) error
	List(ctx context.Context) ([]string, error)
}

// ConfigLoader defines the interface for configuration loading
type ConfigLoader interface {
	Load() error
	GetOdooConfig() OdooConfigProvider
	GetDBConfig() DBConfigProvider
}

// OdooConfigProvider provides Odoo-specific configuration
type OdooConfigProvider interface {
	GetConfigFile() string
	GetOdooBin() string
	GetAddons() []string
}

// DBConfigProvider provides database-specific configuration
type DBConfigProvider interface {
	GetDumpPath() string
	GetDumpFormat() string
}

// ProcessManager defines the interface for process management
type ProcessManager interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	Wait() error
	GetPID() int
}

// Logger defines the interface for logging operations
type Logger interface {
	Info(msg string)
	Error(msg string, err error)
	Warn(msg string)
	Debug(msg string)
}

// SignalHandler defines the interface for signal handling
type SignalHandler interface {
	Register(signals ...os.Signal) <-chan os.Signal
	Unregister(chan<- os.Signal)
}
