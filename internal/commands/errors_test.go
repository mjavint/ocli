package commands

import (
	"bytes"
	"errors"
	"testing"
)

// TestRunEReturnsError verifies commands surface errors through RunE instead of os.Exit/log.Fatal.
func TestRunEReturnsError(t *testing.T) {
	// backupdb requires database name; in the new design this must return an error, not exit.
	var out, errb bytes.Buffer
	deps := Deps{
		Out: &out,
		Err: &errb,
	}
	cmd := NewBackupdbCmd(deps)
	cmd.SetArgs([]string{"--bin", "/bin/true", "--config", "/c", "--database", ""})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("Execute() error = nil, want error for missing database")
	}
	if errors.Is(err, errors.New("unreachable")) {
		t.Fatal("unexpected")
	}
}
