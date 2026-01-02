package config

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Odoo OdooConfig `mapstructure:"odoo"`
	DB   DBSection  `mapstructure:"db"`
}

type OdooConfig struct {
	ConfigFile string   `mapstructure:"config_file"`
	OdooBin    string   `mapstructure:"odoo_bin"`
	Addons     []string `mapstructure:"addons"`
}

type DBSection struct {
	DumpPath   string `mapstructure:"dump_path"`
	DumpFormat string `mapstructure:"dump_format"`
}

type DBConfig struct {
	Host     string
	Port     int
	User     string
	Password string
}

var AppConfig Config

func LoadConfig() {
	// Primero verificar si el archivo de configuración existe
	configFile := "ocli.yml"
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		// El archivo no existe, establecer valores predeterminados
		AppConfig = Config{
			Odoo: OdooConfig{
				ConfigFile: "/workspace/odoo.conf",
				OdooBin:    "/workspace/odoo/odoo-bin",
				Addons:     []string{},
			},
			DB: DBSection{
				DumpPath:   "/workspace/dbs",
				DumpFormat: "zip",
			},
		}
		return
	}

	// El archivo existe, intentar cargarlo
	viper.SetConfigFile(configFile)
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		panic(fmt.Errorf("error leyendo archivo de configuración: %w", err))
	}
	viper.Unmarshal(&AppConfig)
}

// LoadOdooDBParams extrae los parámetros de BD del archivo de configuración
func LoadOdooDBParams(configPath string) (*DBConfig, error) {
	file, err := os.Open(configPath)
	if err != nil {
		return nil, fmt.Errorf("error abriendo %s: %w", configPath, err)
	}
	defer file.Close()

	dbConfig := &DBConfig{}
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Ignorar líneas vacías, comentarios y secciones
		if line == "" ||
			strings.HasPrefix(line, ";") ||
			strings.HasPrefix(line, "#") ||
			(strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]")) {
			continue
		}

		// Procesar clave-valor
		if parts := strings.SplitN(line, "=", 2); len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])

			switch key {
			case "db_host":
				dbConfig.Host = value
			case "db_port":
				port, err := strconv.Atoi(value)
				if err != nil {
					return nil, fmt.Errorf("error convirtiendo db_port a entero: %w", err)
				}
				dbConfig.Port = port
			case "db_user":
				dbConfig.User = value
			case "db_password":
				dbConfig.Password = value
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error leyendo archivo: %w", err)
	}

	// Validar que todos los parámetros necesarios existan
	if dbConfig.Host == "" || dbConfig.Port == 0 ||
		dbConfig.User == "" || dbConfig.Password == "" {
		return nil, errors.New("faltan parámetros de base de datos en el archivo de configuración")
	}

	return dbConfig, nil
}
