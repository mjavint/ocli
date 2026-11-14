package internal

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/mjavint/ocli/internal/commands"
	"github.com/mjavint/ocli/internal/selfupdate"
	"github.com/mjavint/ocli/internal/version"
	"github.com/mjavint/ocli/pkg/config"
)

type countingLoader struct {
	calls int
}

func (c *countingLoader) Load() *config.Config {
	c.calls++
	return &config.Config{}
}

func TestRootVersionFlag(t *testing.T) {
	deps := commands.Deps{
		Version: version.Info{Version: "v1.0.5"},
		Out:     &bytes.Buffer{},
		Err:     &bytes.Buffer{},
		Config:  &countingLoader{},
	}
	cmd := NewRoot(deps)
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"--version"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if got := out.String(); got != "ocli v1.0.5\n" {
		t.Fatalf("output = %q, want %q", got, "ocli v1.0.5\n")
	}
}

func TestVersionCmdRegistered(t *testing.T) {
	deps := commands.Deps{
		Version: version.Info{Version: "v1.0.5", Commit: "23ef66b", Date: "2026-04-19T14:19:35Z"},
		Out:     &bytes.Buffer{},
		Err:     &bytes.Buffer{},
		Config:  &countingLoader{},
	}
	cmd := NewRoot(deps)
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"version"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !strings.Contains(out.String(), "ocli v1.0.5") {
		t.Fatalf("output = %q, want version info", out.String())
	}
}

func TestSelfCommandsDoNotLoadConfig(t *testing.T) {
	loader := &countingLoader{}
	for _, name := range []string{"version", "install", "--version", "--help"} {
		t.Run(name, func(t *testing.T) {
			before := loader.calls
			deps := commands.Deps{
				Version:    version.Info{Version: "v1.0.5"},
				Out:        &bytes.Buffer{},
				Err:        &bytes.Buffer{},
				Config:     loader,
				Executable: func() (string, error) { return "/tmp/ocli", nil },
				Installer:  noopInstaller{},
			}
			cmd := NewRoot(deps)
			cmd.SetOut(&bytes.Buffer{})
			cmd.SetErr(&bytes.Buffer{})
			if name == "install" {
				cmd.SetArgs([]string{"install", "--dir", t.TempDir()})
			} else {
				cmd.SetArgs([]string{name})
			}
			_ = cmd.Execute() // install may fail in RED, but must not load config
			if loader.calls != before {
				t.Fatalf("config loaded for %s: calls %d -> %d", name, before, loader.calls)
			}
		})
	}
}

type noopInstaller struct{}

func (noopInstaller) Install(opts selfupdate.InstallOptions) (selfupdate.InstallResult, error) {
	return selfupdate.InstallResult{Path: opts.Dir + "/ocli", Version: "v1.0.5", SHA256: "x", Changed: false}, nil
}

func TestExecuteReturnsCodes(t *testing.T) {
	// Usage/unknown command should map to exit code 2 in new design.
	code := Execute(context.Background(), []string{"__unknown__"}, &bytes.Buffer{}, &bytes.Buffer{})
	if code != 2 {
		t.Fatalf("Execute(unknown) = %d, want 2", code)
	}
}
