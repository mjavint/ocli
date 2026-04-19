package config

import (
	"github.com/mjavint/ocli/pkg/constants"
)

// Config represents the application configuration
type Config struct {
	Odoo OdooConfig `mapstructure:"odoo"`
	DB   DBSection  `mapstructure:"db"`
}

// OdooConfig represents Odoo-specific configuration
type OdooConfig struct {
	ConfigFile string   `mapstructure:"config_file"`
	OdooBin    string   `mapstructure:"odoo_bin"`
	Addons     []string `mapstructure:"addons"`
}

// DBSection represents database-specific configuration
type DBSection struct {
	DumpPath   string `mapstructure:"dump_path"`
	DumpFormat string `mapstructure:"dump_format"`
}

// DBConfig represents database connection configuration
type DBConfig struct {
	Host     string
	Port     int
	User     string
	Password string
}

// AppConfig is the global configuration instance (for backward compatibility)
var AppConfig Config

// LoadConfig loads the application configuration (legacy function)
// This is maintained for backward compatibility
func LoadConfig() {
	manager := NewConfigManager(constants.DefaultConfigFile)
	if err := manager.Load(); err != nil {
		// Fallback to defaults on error
		AppConfig = *manager.getDefaultConfig()
		return
	}
	if manager.config != nil {
		AppConfig = *manager.config
	}
}

// LoadOdooDBParams extracts database parameters from Odoo config file (legacy function)
// This is maintained for backward compatibility
func LoadOdooDBParams(configPath string) (*DBConfig, error) {
	parser := NewOdooConfigParser()
	return parser.ParseDBParams(configPath)
}
