package internal

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/mjavint/ocli/internal/commands"
	"github.com/mjavint/ocli/internal/odoo"
	"github.com/mjavint/ocli/internal/selfupdate"
	"github.com/mjavint/ocli/internal/version"
	"github.com/mjavint/ocli/pkg/config"
	"github.com/mjavint/ocli/pkg/constants"
	"github.com/spf13/cobra"
)

// ErrUsage indicates invalid CLI usage.
var ErrUsage = errors.New("usage")

// NewRoot creates the root command wired with the provided dependencies.
func NewRoot(deps commands.Deps) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ocli",
		Short: "A brief description of your application",
		Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	if deps.Out != nil {
		cmd.SetOut(deps.Out)
	}
	if deps.Err != nil {
		cmd.SetErr(deps.Err)
	}

	if deps.Config == nil {
		deps.Config = config.NewLoader(constants.DefaultConfigFile)
	}
	if deps.Executable == nil {
		deps.Executable = os.Executable
	}

	cmd.Version = deps.Version.Version
	cmd.SetVersionTemplate("{{.Name}} {{.Version}}\n")

	cmd.AddCommand(commands.NewInitCmd(deps))
	cmd.AddCommand(commands.NewListdbCmd(deps))
	cmd.AddCommand(commands.NewBackupdbCmd(deps))
	cmd.AddCommand(commands.NewRestoredbCmd(deps))
	cmd.AddCommand(commands.NewCopydbCmd(deps))
	cmd.AddCommand(commands.NewInitDBCmd(deps))
	cmd.AddCommand(commands.NewDropdbCmd(deps))
	cmd.AddCommand(commands.NewRenamedbCmd(deps))
	cmd.AddCommand(commands.NewConfigAddonCmd(deps))
	cmd.AddCommand(commands.NewStartOdooCmd(deps))
	cmd.AddCommand(commands.NewVersionCmd(deps))
	cmd.AddCommand(commands.NewInstallCmd(deps))

	for _, c := range cmd.Commands() {
		c.SetOut(nil)
		c.SetErr(nil)
	}

	return cmd
}

// Execute constructs dependencies and runs the CLI.
func Execute(ctx context.Context, args []string, out, errW io.Writer) int {
	info := version.Current()
	deps := commands.Deps{
		Config:     config.NewLoader(constants.DefaultConfigFile),
		Runner:     odoo.NewExecRunner(),
		Installer:  selfupdate.NewInstaller(info),
		Version:    info,
		Executable: os.Executable,
		Out:        out,
		Err:        errW,
	}
	root := NewRoot(deps)
	root.SetArgs(args)
	root.SetOut(out)
	root.SetErr(errW)
	root.SetContext(ctx)

	err := root.ExecuteContext(ctx)
	if err == nil {
		return 0
	}
	if errors.Is(err, ErrUsage) || isUsageError(err) {
		fmt.Fprintf(errW, "error: %v\n", err)
		return 2
	}
	fmt.Fprintf(errW, "error: %v\n", err)
	return 1
}

func isUsageError(err error) bool {
	if err == nil {
		return false
	}
	var flagErr interface{ IsFlagError() bool }
	if errors.As(err, &flagErr) {
		return true
	}
	// cobra unknown command errors
	msg := err.Error()
	return len(msg) >= 15 && msg[:15] == "unknown command"
}
