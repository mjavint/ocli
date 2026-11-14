package commands

import (
	"bytes"
	"context"
	"testing"

	"github.com/mjavint/ocli/pkg/db"
)

func TestListdbOutput(t *testing.T) {
	var out, errBuf bytes.Buffer
	cmd := NewListdbCmd(Deps{
		Out: &out,
		Err: &errBuf,
		ListDatabases: func(ctx context.Context, cfg *db.PGConfig) ([]db.DatabaseSize, error) {
			return []db.DatabaseSize{
				{Name: "alpha", Size: "1.00 KB"},
				{Name: "beta", Size: "2.00 MB"},
			}, nil
		},
	})
	cmd.SetArgs(nil)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v, stderr=%s", err, errBuf.String())
	}
	got := out.String()
	want := "listdb called\n\nFound 2 database(s):\n\n- alpha (1.00 KB)\n- beta (2.00 MB)\n\n"
	if got != want {
		t.Fatalf("output = %q, want %q", got, want)
	}
}
