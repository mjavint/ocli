package commands

import (
	"testing"

	"github.com/mjavint/ocli/pkg/db"
)

func TestOdooStructure(t *testing.T) {
	odoo := &Odoo{
		odooBin:    "/path/to/odoo-bin",
		configPath: "/path/to/odoo.conf",
	}

	if odoo.odooBin != "/path/to/odoo-bin" {
		t.Errorf("odooBin = %v, want /path/to/odoo-bin", odoo.odooBin)
	}

	if odoo.configPath != "/path/to/odoo.conf" {
		t.Errorf("configPath = %v, want /path/to/odoo.conf", odoo.configPath)
	}
}

func TestNewStartOdooCmd(t *testing.T) {
	cmd := NewStartOdooCmd(Deps{})

	if cmd == nil {
		t.Fatal("NewStartOdooCmd() returned nil")
	}

	if cmd.Use != "start" {
		t.Errorf("Use = %v, want start", cmd.Use)
	}

	if cmd.Short != "Odoo Start Server" {
		t.Errorf("Short = %v, want 'Odoo Start Server'", cmd.Short)
	}
}

func TestNewInitCmd(t *testing.T) {
	cmd := NewInitCmd(Deps{})

	if cmd == nil {
		t.Fatal("NewInitCmd() returned nil")
	}

	if cmd.Use != "init" {
		t.Errorf("Use = %v, want init", cmd.Use)
	}
}

func TestNewDropdbCmd(t *testing.T) {
	cmd := NewDropdbCmd(Deps{})

	if cmd == nil {
		t.Fatal("NewDropdbCmd() returned nil")
	}

	if cmd.Use != "dropdb" {
		t.Errorf("Use = %v, want dropdb", cmd.Use)
	}
}

func TestNewListdbCmd(t *testing.T) {
	cmd := NewListdbCmd(Deps{})

	if cmd == nil {
		t.Fatal("NewListdbCmd() returned nil")
	}

	if cmd.Use != "listdb" {
		t.Errorf("Use = %v, want listdb", cmd.Use)
	}
}

func TestNewInitDBCmd(t *testing.T) {
	cmd := NewInitDBCmd(Deps{})

	if cmd == nil {
		t.Fatal("NewInitDBCmd() returned nil")
	}

	if cmd.Use != "initdb" {
		t.Errorf("Use = %v, want initdb", cmd.Use)
	}
}

func TestNewCopydbCmd(t *testing.T) {
	cmd := NewCopydbCmd(Deps{})

	if cmd == nil {
		t.Fatal("NewCopydbCmd() returned nil")
	}

	if cmd.Use != "copydb" {
		t.Errorf("Use = %v, want copydb", cmd.Use)
	}
}

func TestNewRenamedbCmd(t *testing.T) {
	cmd := NewRenamedbCmd(Deps{})

	if cmd == nil {
		t.Fatal("NewRenamedbCmd() returned nil")
	}

	if cmd.Use != "renamedb" {
		t.Errorf("Use = %v, want renamedb", cmd.Use)
	}
}

func TestNewBackupdbCmd(t *testing.T) {
	cmd := NewBackupdbCmd(Deps{})

	if cmd == nil {
		t.Fatal("NewBackupdbCmd() returned nil")
	}

	if cmd.Use != "backupdb" {
		t.Errorf("Use = %v, want backupdb", cmd.Use)
	}
}

func TestNewRestoredbCmd(t *testing.T) {
	cmd := NewRestoredbCmd(Deps{})

	if cmd == nil {
		t.Fatal("NewRestoredbCmd() returned nil")
	}

	if cmd.Use != "restoredb" {
		t.Errorf("Use = %v, want restoredb", cmd.Use)
	}
}

func TestNewConfigAddonCmd(t *testing.T) {
	cmd := NewConfigAddonCmd(Deps{})

	if cmd == nil {
		t.Fatal("NewConfigAddonCmd() returned nil")
	}

	if cmd.Use != "addon" {
		t.Errorf("Use = %v, want addon", cmd.Use)
	}
}

func TestIsValidDBNameWrapper(t *testing.T) {
	// Test the db.IsValidDBName function through commands package
	tests := []struct {
		name   string
		dbname string
		want   bool
	}{
		{"valid name", "test_db", true},
		{"invalid name", "test-db", false},
		{"empty name", "", false},
		{"starts with number", "1testdb", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := db.IsValidDBName(tt.dbname)
			if got != tt.want {
				t.Errorf("IsValidDBName(%q) = %v, want %v", tt.dbname, got, tt.want)
			}
		})
	}
}
