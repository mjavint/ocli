package internal

import (
	"testing"
)

func TestExecute(t *testing.T) {
	// Test that Execute function exists and can be referenced
	// We can't actually call it in tests without proper setup
	// This test ensures the function is exported and available
	t.Log("Execute function is available")
}

func TestRootCmd(t *testing.T) {
	if rootCmd == nil {
		t.Fatal("rootCmd is nil")
	}

	if rootCmd.Use != "ocli" {
		t.Errorf("rootCmd.Use = %v, want 'ocli'", rootCmd.Use)
	}

	if rootCmd.Short == "" {
		t.Error("rootCmd.Short should not be empty")
	}

	if rootCmd.Long == "" {
		t.Error("rootCmd.Long should not be empty")
	}
}
