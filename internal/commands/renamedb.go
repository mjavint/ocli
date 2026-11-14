package commands

import (
	"fmt"

	"github.com/mjavint/ocli/internal/odoo"
	"github.com/spf13/cobra"
)

// NewRenamedbCmd creates the renamedb command.
func NewRenamedbCmd(deps Deps) *cobra.Command {
	var f odooDBFlags
	cmd := &cobra.Command{
		Use:   "renamedb",
		Short: "Rename an Odoo database",
		Long: `Rename an Odoo database from one name to another.
This command connects to the Odoo database management API
and performs a database rename operation.

Example:
  ocli renamedb mydb_old mydb_new`,
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			fmt.Fprintln(out, "db called")
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
			if f.NewDB == "" {
				return fmt.Errorf("New database name is required. Use --new-db or -n to specify it.")
			}

			f.ForceSet = cmd.Flags().Changed("force")
			cmdArgs := buildRenameArgs(f)
			runner := deps.Runner
			if runner == nil {
				runner = odoo.NewExecRunner()
			}
			if err := runner.Run(cmd.Context(), f.Bin, cmdArgs, odoo.RunOptions{Stdout: deps.stdout(), Stderr: deps.stderr()}); err != nil {
				return fmt.Errorf("Rename failed: %v", err)
			}
			fmt.Fprintf(out, "Rename completed successfully: %s\n", f.NewDB)
			return nil
		},
	}
	cmd.Flags().StringVarP(&f.Bin, "bin", "b", "", "Path to the Odoo binary")
	cmd.Flags().StringVarP(&f.Config, "config", "c", "", "Odoo configuration file path (odoo.conf)")
	cmd.Flags().StringVarP(&f.NewDB, "new-db", "n", "", "New database name to restore to")
	cmd.Flags().StringVarP(&f.Database, "database", "d", "", "Database name to backup")
	cmd.Flags().BoolVarP(&f.Force, "force", "f", true, "Force restore even if the database already exists")
	setCmdIO(cmd, deps)
	return cmd
}
