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
			got, err := GetBackupFilePath(tt.backupDir, tt.dbName, tt.backupFormat, tt.noFilestore)
			if err != nil {
				t.Fatalf("GetBackupFilePath() error = %v", err)
			}
			if got != tt.want {
				t.Errorf("GetBackupFilePath() = %v, want %v", got, tt.want)
			}

			// Verify directory was created
			if _, err := os.Stat(tt.backupDir); os.IsNotExist(err) {
				t.Errorf("GetBackupFilePath() did not create directory: %v", tt.backupDir)
			}
		})
	}
}

func TestGetBackupFilePathCreatesDirectory(t *testing.T) {
	tempDir := t.TempDir()
	newDir := filepath.Join(tempDir, "backups", "nested")

	path, err := GetBackupFilePath(newDir, "test_db", "sql", false)
	if err != nil {
		t.Fatalf("GetBackupFilePath() error = %v", err)
	}

	if path == "" {
		t.Error("GetBackupFilePath() returned empty string")
	}

	if _, err := os.Stat(newDir); os.IsNotExist(err) {
		t.Errorf("GetBackupFilePath() did not create nested directory: %v", newDir)
	}

	expectedPath := filepath.Join(newDir, "test_db.sql")
	if path != expectedPath {
		t.Errorf("GetBackupFilePath() = %v, want %v", path, expectedPath)
	}
}

func TestGetBackupFilePathError(t *testing.T) {
	tempDir := t.TempDir()
	// Create a file where a directory is expected to force MkdirAll failure.
	blocker := filepath.Join(tempDir, "not-a-dir")
	if err := os.WriteFile(blocker, []byte("x"), 0o600); err != nil {
		t.Fatalf("write blocker: %v", err)
	}

	_, err := GetBackupFilePath(filepath.Join(blocker, "child"), "db", "sql", false)
	if err == nil {
		t.Fatal("GetBackupFilePath() error = nil, want error")
	}
}
