package commands

import (
	"fmt"
	"os"

	"github.com/mjavint/ocli/internal/selfupdate"
	"github.com/spf13/cobra"
)

// NewInstallCmd creates the install command.
func NewInstallCmd(deps Deps) *cobra.Command {
	var (
		dir   string
		force bool
	)
	cmd := &cobra.Command{
		Use:   "install",
		Short: "Install the ocli binary into a directory on the PATH",
		Long: `Copy the running ocli binary into a directory on the PATH.

The default destination is ~/.local/bin. Use --dir to choose another
directory. Re-running install is idempotent: when the destination already
contains the same binary, nothing is written.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			execPath := ""
			if deps.Executable != nil {
				p, err := deps.Executable()
				if err != nil {
					return err
				}
				execPath = p
			} else {
				p, err := os.Executable()
				if err != nil {
					return err
				}
				execPath = p
			}

			installer := deps.Installer
			if installer == nil {
				installer = selfupdate.NewInstaller(deps.Version)
			}

			res, err := installer.Install(selfupdate.InstallOptions{Source: execPath, Dir: dir, Force: force})
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			errOut := cmd.ErrOrStderr()
			if res.Changed {
				fmt.Fprintf(out, "installed ocli %s to %s (sha256: %s)\n", res.Version, res.Path, res.SHA256)
			} else {
				fmt.Fprintf(out, "ocli %s already installed at %s (sha256: %s)\n", res.Version, res.Path, res.SHA256)
			}
			if res.PathWarning != "" {
				fmt.Fprintln(errOut, res.PathWarning)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&dir, "dir", "", "Directory to install ocli into (default ~/.local/bin)")
	cmd.Flags().BoolVar(&force, "force", false, "Replace the destination even if it is already up to date")
	setCmdIO(cmd, deps)
	return cmd
}
