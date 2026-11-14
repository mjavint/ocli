package commands

import (
	"fmt"

	"github.com/mjavint/ocli/internal/odoo"
	"github.com/spf13/cobra"
)

// NewDropdbCmd creates the dropdb command.
func NewDropdbCmd(deps Deps) *cobra.Command {
	var f odooDBFlags
	cmd := &cobra.Command{
		Use:   "dropdb",
		Short: "A brief description of your command",
		Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			fmt.Fprintln(out, "dropdb called")
			cfg := deps.cfg()
			if f.Bin == "" {
				f.Bin = cfg.Odoo.OdooBin
			}
			if f.Config == "" {
				f.Config = cfg.Odoo.ConfigFile
			}
			if f.Database == "" {
				return fmt.Errorf("Database name is required. Use --database or -d to specify it.")
			}

			cmdArgs := buildDropArgs(f)
			runner := deps.Runner
			if runner == nil {
				runner = odoo.NewExecRunner()
			}
			output, err := runner.CombinedOutput(cmd.Context(), f.Bin, cmdArgs)
			if err != nil {
				return fmt.Errorf("Error executing drop command: %v\nOutput: %s", err, string(output))
			}
			fmt.Fprintf(out, "Database drop completed successfully: %s\n", f.Database)
			return nil
		},
	}
	cmd.Flags().StringVarP(&f.Bin, "bin", "b", "", "Path to the Odoo binary")
	cmd.Flags().StringVarP(&f.Config, "config", "c", "", "Odoo configuration file path (odoo.conf)")
	cmd.Flags().StringVarP(&f.Database, "database", "d", "", "Database name to backup")
	setCmdIO(cmd, deps)
	return cmd
}
