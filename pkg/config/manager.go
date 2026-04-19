package config

import (
	"fmt"
	"os"

	"github.com/mjavint/ocli/pkg/constants"
	apperrors "github.com/mjavint/ocli/pkg/errors"
	"github.com/spf13/viper"
)

// ConfigManager handles configuration loading and management
type ConfigManager struct {
	config     *Config
	configFile string
}

// NewConfigManager creates a new configuration manager
func NewConfigManager(configFile string) *ConfigManager {
	if configFile == "" {
		configFile = constants.DefaultConfigFile
	}
	return &ConfigManager{
		configFile: configFile,
	}
}

// Load loads the configuration from file or uses defaults
func (cm *ConfigManager) Load() error {
	// Check if config file exists
	if _, err := os.Stat(cm.configFile); os.IsNotExist(err) {
		cm.config = cm.getDefaultConfig()
		return nil
	}

	// Load from file
	config, err := cm.loadFromFile()
	if err != nil {
		return apperrors.NewInvalidConfigError("failed to load config file", err)
	}

	cm.config = config
	return nil
}

// loadFromFile loads configuration from a file
func (cm *ConfigManager) loadFromFile() (*Config, error) {
	v := viper.New()
	v.SetConfigFile(cm.configFile)
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	var config Config
	if err := v.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	return &config, nil
}

// getDefaultConfig returns default configuration
func (cm *ConfigManager) getDefaultConfig() *Config {
	return &Config{
		Odoo: OdooConfig{
			ConfigFile: constants.DefaultOdooConfigFile,
			OdooBin:    constants.DefaultOdooBin,
			Addons:     []string{},
		},
		DB: DBSection{
			DumpPath:   constants.DefaultDumpPath,
			DumpFormat: constants.DefaultDumpFormat,
		},
	}
}

// GetConfig returns the loaded configuration
func (cm *ConfigManager) GetConfig() *Config {
	return cm.config
}

// GetOdooConfig returns the Odoo configuration
func (cm *ConfigManager) GetOdooConfig() OdooConfigProvider {
	if cm.config != nil {
		return &cm.config.Odoo
	}
	return nil
}

// GetDBConfig returns the database configuration
func (cm *ConfigManager) GetDBConfig() DBConfigProvider {
	if cm.config != nil {
		return &cm.config.DB
	}
	return nil
}

// OdooConfigProvider implementation
type OdooConfigProvider interface {
	GetConfigFile() string
	GetOdooBin() string
	GetAddons() []string
}

// DBConfigProvider implementation
type DBConfigProvider interface {
	GetDumpPath() string
	GetDumpFormat() string
}

// GetConfigFile implements OdooConfigProvider
func (o *OdooConfig) GetConfigFile() string {
	return o.ConfigFile
}

// GetOdooBin implements OdooConfigProvider
func (o *OdooConfig) GetOdooBin() string {
	return o.OdooBin
}

// GetAddons implements OdooConfigProvider
func (o *OdooConfig) GetAddons() []string {
	return o.Addons
}

// GetDumpPath implements DBConfigProvider
func (d *DBSection) GetDumpPath() string {
	return d.DumpPath
}

// GetDumpFormat implements DBConfigProvider
func (d *DBSection) GetDumpFormat() string {
	return d.DumpFormat
}
