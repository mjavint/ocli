package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadOdooDBParams(t *testing.T) {
	// Create a temporary config file
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "odoo.conf")

	configContent := `[options]
db_host = localhost
db_port = 5432
db_user = odoo
db_password = odoopass
db_name = odoodb
admin_passwd = admin123
`

	err := os.WriteFile(configFile, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	// Test loading config
	dbConfig, err := LoadOdooDBParams(configFile)
	if err != nil {
		t.Fatalf("LoadOdooDBParams() error = %v", err)
	}

	if dbConfig == nil {
		t.Fatal("LoadOdooDBParams() returned nil config")
	}

	// Verify values
	tests := []struct {
		name string
		got  interface{}
		want interface{}
	}{
		{"Host", dbConfig.Host, "localhost"},
		{"Port", dbConfig.Port, 5432},
		{"User", dbConfig.User, "odoo"},
		{"Password", dbConfig.Password, "odoopass"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("%s = %v, want %v", tt.name, tt.got, tt.want)
			}
		})
	}
}

func TestLoadOdooDBParamsInvalidFile(t *testing.T) {
	_, err := LoadOdooDBParams("/nonexistent/file.conf")
	if err == nil {
		t.Error("LoadOdooDBParams() should return error for nonexistent file")
	}
}

func TestLoadOdooDBParamsInvalidPort(t *testing.T) {
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "odoo.conf")

	configContent := `[options]
db_host = localhost
db_port = invalid_port
db_user = odoo
db_password = pass
`

	err := os.WriteFile(configFile, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	_, err = LoadOdooDBParams(configFile)
	if err == nil {
		t.Error("LoadOdooDBParams() should return error for invalid port")
	}
}

func TestLoadOdooDBParamsPartialConfig(t *testing.T) {
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "odoo.conf")

	// Incomplete config - missing required fields
	configContent := `[options]
db_host = testhost
`

	err := os.WriteFile(configFile, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	_, err = LoadOdooDBParams(configFile)
	if err == nil {
		t.Error("LoadOdooDBParams() should return error for incomplete config")
	}

	if err != nil && !strings.Contains(err.Error(), "faltan parámetros") {
		t.Errorf("Expected missing parameters error, got: %v", err)
	}
}

func TestLoadOdooDBParamsWithComments(t *testing.T) {
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "odoo.conf")

	configContent := `# This is a comment
[options]
# db_host = commented_out
db_host = actualhost
db_port = 5433
; Another comment style
db_user = testuser
db_password = testpass
`

	err := os.WriteFile(configFile, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	dbConfig, err := LoadOdooDBParams(configFile)
	if err != nil {
		t.Fatalf("LoadOdooDBParams() error = %v", err)
	}

	if dbConfig.Host != "actualhost" {
		t.Errorf("Host = %v, want actualhost", dbConfig.Host)
	}

	if dbConfig.Port != 5433 {
		t.Errorf("Port = %v, want 5433", dbConfig.Port)
	}
}

func TestDBConfigDefaults(t *testing.T) {
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "odoo.conf")

	// Empty config file - should fail validation
	configContent := `[options]
`

	err := os.WriteFile(configFile, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	_, err = LoadOdooDBParams(configFile)
	if err == nil {
		t.Error("LoadOdooDBParams() should return error for empty config")
	}

	if err != nil && !strings.Contains(err.Error(), "faltan parámetros") {
		t.Errorf("Expected missing parameters error, got: %v", err)
	}
}
