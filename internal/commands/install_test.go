package commands

import (
	"bytes"
	"strings"
	"testing"

	"github.com/mjavint/ocli/internal/selfupdate"
	"github.com/mjavint/ocli/internal/version"
)

type fakeInstaller struct {
	called bool
	opts   selfupdate.InstallOptions
	res    selfupdate.InstallResult
	err    error
}

func (f *fakeInstaller) Install(opts selfupdate.InstallOptions) (selfupdate.InstallResult, error) {
	f.called = true
	f.opts = opts
	return f.res, f.err
}

func TestInstallCmd(t *testing.T) {
	var out, errb bytes.Buffer
	fi := &fakeInstaller{
		res: selfupdate.InstallResult{Path: "/home/u/.local/bin/ocli", Version: "v1.0.5", SHA256: "abc", Changed: true},
	}
	deps := Deps{
		Installer:  fi,
		Version:    version.Info{Version: "v1.0.5"},
		Executable: func() (string, error) { return "/tmp/build/ocli", nil },
		Out:        &out,
		Err:        &errb,
	}
	cmd := NewInstallCmd(deps)
	cmd.SetArgs([]string{"--dir", "/home/u/.local/bin"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !fi.called {
		t.Fatal("Install() was not called")
	}
	if fi.opts.Dir != "/home/u/.local/bin" {
		t.Fatalf("Dir = %q, want /home/u/.local/bin", fi.opts.Dir)
	}
	if fi.opts.Source != "/tmp/build/ocli" {
		t.Fatalf("Source = %q, want /tmp/build/ocli", fi.opts.Source)
	}
	if !strings.Contains(out.String(), "installed ocli v1.0.5") {
		t.Fatalf("output = %q, want installed message", out.String())
	}
}

func TestInstallCmdNoop(t *testing.T) {
	var out bytes.Buffer
	fi := &fakeInstaller{
		res: selfupdate.InstallResult{Path: "/home/u/.local/bin/ocli", Version: "v1.0.5", SHA256: "abc", Changed: false},
	}
	deps := Deps{
		Installer:  fi,
		Version:    version.Info{Version: "v1.0.5"},
		Executable: func() (string, error) { return "/tmp/build/ocli", nil },
		Out:        &out,
		Err:        &bytes.Buffer{},
	}
	cmd := NewInstallCmd(deps)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !strings.Contains(out.String(), "already installed") {
		t.Fatalf("output = %q, want already installed message", out.String())
	}
}

func TestInstallCmdForceFlag(t *testing.T) {
	fi := &fakeInstaller{res: selfupdate.InstallResult{Changed: true, Path: "/x/ocli", Version: "v1.0.5", SHA256: "a"}}
	deps := Deps{
		Installer:  fi,
		Version:    version.Info{Version: "v1.0.5"},
		Executable: func() (string, error) { return "/tmp/ocli", nil },
		Out:        &bytes.Buffer{},
		Err:        &bytes.Buffer{},
	}
	cmd := NewInstallCmd(deps)
	cmd.SetArgs([]string{"--force"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !fi.opts.Force {
		t.Fatal("Force = false, want true")
	}
}
