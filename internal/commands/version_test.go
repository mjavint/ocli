package commands

import (
	"bytes"
	"strings"
	"testing"

	"github.com/mjavint/ocli/internal/version"
)

func TestVersionCmd(t *testing.T) {
	var out, errb bytes.Buffer
	deps := Deps{
		Version: version.Info{Version: "v1.0.5", Commit: "23ef66b", Date: "2026-04-19T14:19:35Z"},
		Out:     &out,
		Err:     &errb,
	}
	cmd := NewVersionCmd(deps)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	want := "ocli v1.0.5 (commit 23ef66b, built 2026-04-19T14:19:35Z)\n"
	if got := out.String(); got != want {
		t.Fatalf("output = %q, want %q", got, want)
	}
	if errb.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", errb.String())
	}
}

func TestVersionCmdDirty(t *testing.T) {
	var out bytes.Buffer
	deps := Deps{
		Version: version.Info{Version: "v1.0.5", Commit: "23ef66b", Date: "2026-04-19T14:19:35Z", Modified: true},
		Out:     &out,
		Err:     &bytes.Buffer{},
	}
	cmd := NewVersionCmd(deps)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !strings.Contains(out.String(), "[dirty]") {
		t.Fatalf("output = %q, want [dirty]", out.String())
	}
}
