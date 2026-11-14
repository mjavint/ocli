package commands

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/mjavint/ocli/internal/selfupdate"
	"github.com/mjavint/ocli/internal/version"
)

type fakeUpdater struct {
	opts selfupdate.UpdateOptions
	res  selfupdate.UpdateResult
	err  error
}

func (f *fakeUpdater) Update(ctx context.Context, opts selfupdate.UpdateOptions) (selfupdate.UpdateResult, error) {
	f.opts = opts
	return f.res, f.err
}

func TestUpdateCheckCmd(t *testing.T) {
	var out, errb bytes.Buffer
	fu := &fakeUpdater{
		res: selfupdate.UpdateResult{Current: "v1.0.4", Latest: "v1.0.5", Available: true},
		err: selfupdate.ErrUpdateAvailable,
	}
	deps := Deps{
		Updater:    fu,
		Version:    version.Info{Version: "v1.0.4"},
		Executable: func() (string, error) { return "/tmp/ocli", nil },
		Out:        &out,
		Err:        &errb,
	}
	cmd := NewUpdateCmd(deps)
	cmd.SetArgs([]string{"--check"})
	err := cmd.Execute()
	if !errors.Is(err, selfupdate.ErrUpdateAvailable) {
		t.Fatalf("Execute() err = %v, want ErrUpdateAvailable", err)
	}
	if !fu.opts.CheckOnly {
		t.Fatal("CheckOnly = false, want true")
	}
	if !strings.Contains(out.String(), "update available") {
		t.Fatalf("output = %q, want update available", out.String())
	}
}

func TestUpdateCmdUpToDate(t *testing.T) {
	var out bytes.Buffer
	fu := &fakeUpdater{res: selfupdate.UpdateResult{Current: "v1.0.5", Latest: "v1.0.5", SHA256: "abc", Updated: false}}
	deps := Deps{
		Updater:    fu,
		Version:    version.Info{Version: "v1.0.5"},
		Executable: func() (string, error) { return "/tmp/ocli", nil },
		Out:        &out,
		Err:        &bytes.Buffer{},
	}
	cmd := NewUpdateCmd(deps)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !strings.Contains(out.String(), "up to date") {
		t.Fatalf("output = %q, want up to date", out.String())
	}
}

func TestUpdateCmdApplies(t *testing.T) {
	var out bytes.Buffer
	fu := &fakeUpdater{res: selfupdate.UpdateResult{Current: "v1.0.4", Latest: "v1.0.5", SHA256: "abc", Updated: true}}
	deps := Deps{
		Updater:    fu,
		Version:    version.Info{Version: "v1.0.4"},
		Executable: func() (string, error) { return "/tmp/ocli", nil },
		Out:        &out,
		Err:        &bytes.Buffer{},
	}
	cmd := NewUpdateCmd(deps)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !strings.Contains(out.String(), "updated ocli") {
		t.Fatalf("output = %q, want updated message", out.String())
	}
}

func TestUpdateCmdFlags(t *testing.T) {
	fu := &fakeUpdater{res: selfupdate.UpdateResult{Current: "v1.0.4", Latest: "v1.0.5", Updated: true}}
	deps := Deps{
		Updater:    fu,
		Version:    version.Info{Version: "v1.0.4"},
		Executable: func() (string, error) { return "/tmp/ocli", nil },
		Out:        &bytes.Buffer{},
		Err:        &bytes.Buffer{},
	}
	cmd := NewUpdateCmd(deps)
	cmd.SetArgs([]string{"--force", "--version", "v1.0.4", "--checksum", strings.Repeat("a", 64)})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !fu.opts.Force || fu.opts.TargetTag != "v1.0.4" || fu.opts.PinnedSHA256 != strings.Repeat("a", 64) {
		t.Fatalf("opts = %+v, flags not propagated", fu.opts)
	}
}
