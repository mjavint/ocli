package constants

import "time"

// Database constants
const (
	// MaxDBNameLength is the maximum allowed length for a database name
	MaxDBNameLength = 63

	// DefaultDBPort is the default PostgreSQL port
	DefaultDBPort = 5432

	// DefaultDBHost is the default PostgreSQL host
	DefaultDBHost = "localhost"

	// DefaultDBUser is the default PostgreSQL user
	DefaultDBUser = "postgres"

	// DefaultSSLMode is the default SSL mode for PostgreSQL
	DefaultSSLMode = "disable"
)

// Connection pool constants
const (
	// DefaultMaxOpenConns is the default maximum number of open connections
	DefaultMaxOpenConns = 25

	// DefaultMaxIdleConns is the default maximum number of idle connections
	DefaultMaxIdleConns = 5

	// DefaultConnMaxLifetime is the default maximum connection lifetime
	DefaultConnMaxLifetime = 5 * time.Minute

	// DefaultConnMaxIdleTime is the default maximum connection idle time
	DefaultConnMaxIdleTime = 1 * time.Minute
)

// Process constants
const (
	// DefaultShutdownTimeout is the default timeout for graceful shutdown
	DefaultShutdownTimeout = 10 * time.Second

	// SignalBufferSize is the size of the signal channel buffer
	SignalBufferSize = 1
)

// File constants
const (
	// DefaultConfigFile is the default configuration file name
	DefaultConfigFile = "ocli.yml"

	// DefaultOdooConfigFile is the default Odoo configuration file
	DefaultOdooConfigFile = "/workspace/odoo.conf"

	// DefaultOdooBin is the default Odoo binary path
	DefaultOdooBin = "/workspace/odoo/odoo-bin"

	// DefaultDumpPath is the default path for database dumps
	DefaultDumpPath = "/workspace/dbs"

	// DefaultDumpFormat is the default dump format
	DefaultDumpFormat = "zip"
)

// Validation patterns
const (
	// DBNamePattern is the regex pattern for valid database names
	DBNamePattern = `^[a-zA-Z_][a-zA-Z0-9_]*$`
)
