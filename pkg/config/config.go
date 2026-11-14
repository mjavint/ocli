package config

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

// LoadOdooDBParams extracts database parameters from Odoo config file (legacy function)
// This is maintained for backward compatibility
func LoadOdooDBParams(configPath string) (*DBConfig, error) {
	parser := NewOdooConfigParser()
	return parser.ParseDBParams(configPath)
}
