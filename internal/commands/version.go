package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

// NewVersionCmd creates the version command.
func NewVersionCmd(deps Deps) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		RunE: func(cmd *cobra.Command, args []string) error {
			_, err := fmt.Fprintln(cmd.OutOrStdout(), deps.Version.Full())
			return err
		},
	}
	setCmdIO(cmd, deps)
	return cmd
}
