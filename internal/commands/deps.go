package commands

import (
	"io"
	"os"

	"github.com/mjavint/ocli/internal/odoo"
	"github.com/mjavint/ocli/internal/selfupdate"
	"github.com/mjavint/ocli/internal/version"
	"github.com/mjavint/ocli/pkg/config"
	"github.com/spf13/cobra"
)

// ConfigLoader loads the ocli configuration.
type ConfigLoader interface {
	Load() *config.Config
}

// Installer installs the ocli binary on the PATH.
type Installer interface {
	Install(selfupdate.InstallOptions) (selfupdate.InstallResult, error)
}

// Deps holds the dependencies injected into every command.
type Deps struct {
	Config     ConfigLoader
	Runner     odoo.Runner
	Installer  Installer
	Version    version.Info
	Executable func() (string, error)
	Out        io.Writer
	Err        io.Writer
}

func (d Deps) stdout() io.Writer {
	if d.Out != nil {
		return d.Out
	}
	return os.Stdout
}

func (d Deps) stderr() io.Writer {
	if d.Err != nil {
		return d.Err
	}
	return os.Stderr
}

func setCmdIO(cmd *cobra.Command, deps Deps) {
	if deps.Out != nil {
		cmd.SetOut(deps.Out)
	}
	if deps.Err != nil {
		cmd.SetErr(deps.Err)
	}
}

func (d Deps) cfg() *config.Config {
	if d.Config != nil {
		if c := d.Config.Load(); c != nil {
			return c
		}
	}
	return &config.Config{}
}
