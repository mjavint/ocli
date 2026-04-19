package config

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/mjavint/ocli/pkg/constants"
	apperrors "github.com/mjavint/ocli/pkg/errors"
)

// OdooConfigParser handles parsing of Odoo configuration files
type OdooConfigParser struct{}

// NewOdooConfigParser creates a new Odoo config parser
func NewOdooConfigParser() *OdooConfigParser {
	return &OdooConfigParser{}
}

// ParseDBParams extracts database parameters from Odoo config file
func (p *OdooConfigParser) ParseDBParams(configPath string) (*DBConfig, error) {
	if configPath == "" {
		return nil, apperrors.NewInvalidConfigError("config path cannot be empty", nil)
	}

	file, err := os.Open(configPath)
	if err != nil {
		return nil, apperrors.NewInvalidConfigError(
			fmt.Sprintf("failed to open config file: %s", configPath),
			err,
		)
	}
	defer func() {
		if err := file.Close(); err != nil {
			log.Printf("error closing file: %v", err)
		}
	}()

	dbConfig := &DBConfig{
		Host: constants.DefaultDBHost,
		Port: constants.DefaultDBPort,
		User: "odoo",
	}

	scanner := bufio.NewScanner(file)
	if err := p.scanConfigFile(scanner, dbConfig); err != nil {
		return nil, err
	}

	if err := scanner.Err(); err != nil {
		return nil, apperrors.NewInvalidConfigError("error reading config file", err)
	}

	if err := p.validateDBConfig(dbConfig); err != nil {
		return nil, err
	}

	return dbConfig, nil
}

// scanConfigFile scans the config file and populates dbConfig
func (p *OdooConfigParser) scanConfigFile(scanner *bufio.Scanner, dbConfig *DBConfig) error {
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines, comments, and sections
		if p.shouldSkipLine(line) {
			continue
		}

		// Process key-value pairs
		if err := p.processConfigLine(line, dbConfig); err != nil {
			return err
		}
	}
	return nil
}

// shouldSkipLine determines if a line should be skipped
func (p *OdooConfigParser) shouldSkipLine(line string) bool {
	return line == "" ||
		strings.HasPrefix(line, ";") ||
		strings.HasPrefix(line, "#") ||
		(strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]"))
}

// processConfigLine processes a configuration line
func (p *OdooConfigParser) processConfigLine(line string, dbConfig *DBConfig) error {
	parts := strings.SplitN(line, "=", 2)
	if len(parts) != 2 {
		return nil
	}

	key := strings.TrimSpace(parts[0])
	value := strings.TrimSpace(parts[1])

	switch key {
	case "db_host":
		dbConfig.Host = value
	case "db_port":
		port, err := strconv.Atoi(value)
		if err != nil {
			return apperrors.NewInvalidConfigError(
				fmt.Sprintf("invalid port value: %s", value),
				err,
			)
		}
		dbConfig.Port = port
	case "db_user":
		dbConfig.User = value
	case "db_password":
		dbConfig.Password = value
	}

	return nil
}

// validateDBConfig validates the database configuration
func (p *OdooConfigParser) validateDBConfig(dbConfig *DBConfig) error {
	if dbConfig.Host == "" || dbConfig.Port == 0 ||
		dbConfig.User == "" || dbConfig.Password == "" {
		return errors.New("faltan parámetros de base de datos en el archivo de configuración")
	}
	return nil
}
