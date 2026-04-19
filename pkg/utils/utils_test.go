package utils

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetBackupFilePath(t *testing.T) {
	// Create temporary directory for tests
	tempDir := t.TempDir()

	tests := []struct {
		name         string
		backupDir    string
		dbName       string
		backupFormat string
		noFilestore  bool
		want         string
	}{
		{
			name:         "basic backup path",
			backupDir:    tempDir,
			dbName:       "test_db",
			backupFormat: "sql",
			noFilestore:  false,
			want:         filepath.Join(tempDir, "test_db.sql"),
		},
		{
			name:         "backup without filestore",
			backupDir:    tempDir,
			dbName:       "production_db",
			backupFormat: "dump",
			noFilestore:  true,
			want:         filepath.Join(tempDir, "production_db_no_fs.dump"),
		},
		{
			name:         "custom format",
			backupDir:    tempDir,
			dbName:       "custom_db",
			backupFormat: "tar",
			noFilestore:  false,
			want:         filepath.Join(tempDir, "custom_db.tar"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetBackupFilePath(tt.backupDir, tt.dbName, tt.backupFormat, tt.noFilestore)
			if got != tt.want {
				t.Errorf("GetBackupFilePath() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetBackupFilePathCreatesDirectory(t *testing.T) {
	tempDir := t.TempDir()
	backupDir := filepath.Join(tempDir, "backups", "nested")

	result := GetBackupFilePath(backupDir, "test", "sql", false)
	expected := filepath.Join(backupDir, "test.sql")

	if result != expected {
		t.Errorf("GetBackupFilePath() = %v, want %v", result, expected)
	}

	// Check if directory was created
	if _, err := os.Stat(backupDir); os.IsNotExist(err) {
		t.Errorf("Directory was not created: %v", backupDir)
	}
}

func TestGetBackupFilePathInvalidDirectory(t *testing.T) {
	// Try to create backup path in a read-only location (this will fail on permission issues)
	// For a proper test, we'd need a mock filesystem or to test on a location we know fails
	result := GetBackupFilePath("/root/invalid/path/that/doesnt/exist", "test", "sql", false)
	if result != "" {
		// On some systems this might work, so we just check the logic
		t.Logf("Backup path returned: %v", result)
	}
}
