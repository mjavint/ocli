package commands

import (
	"fmt"

	"github.com/mjavint/ocli/internal/odoo"
	"github.com/spf13/cobra"
)

// NewCopydbCmd creates the copydb command.
func NewCopydbCmd(deps Deps) *cobra.Command {
	var f odooDBFlags
	cmd := &cobra.Command{
		Use:   "copydb",
		Short: "Duplicate an Odoo database",
		Long: `Duplicate an existing Odoo database to a new database.

This command creates a complete copy of an Odoo database, including all data,
configurations, and filestore. This is useful for:

- Creating staging or testing environments
- Making backups before major changes
- Setting up development databases from production data

Example:
  ocli copydb source_db target_db
  ocli copydb production_db staging_db

The command will copy the database schema, all tables and data, and optionally
the filestore directory associated with the database.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			fmt.Fprintln(out, "copydb called")
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
			f.NeutralizeSet = cmd.Flags().Changed("neutralize")
			cmdArgs := buildCopyArgs(f)
			fmt.Fprintf(out, "Executing: %s %v\n", f.Bin, cmdArgs)

			runner := deps.Runner
			if runner == nil {
				runner = odoo.NewExecRunner()
			}
			if err := runner.Run(cmd.Context(), f.Bin, cmdArgs, odoo.RunOptions{Stdout: out, Stderr: cmd.ErrOrStderr()}); err != nil {
				return fmt.Errorf("Error executing duplicate command: %v", err)
			}
			fmt.Fprintf(out, "Duplicate completed successfully: %s\n", f.NewDB)
			return nil
		},
	}
	cmd.Flags().StringVarP(&f.Bin, "bin", "b", "", "Path to the Odoo binary")
	cmd.Flags().StringVarP(&f.Config, "config", "c", "", "Odoo configuration file path (odoo.conf)")
	cmd.Flags().StringVarP(&f.Database, "database", "d", "", "Database name to backup")
	cmd.Flags().StringVarP(&f.NewDB, "new-db", "n", "", "New database name to restore to")
	cmd.Flags().BoolVarP(&f.Force, "force", "f", true, "Force restore even if the database already exists")
	cmd.Flags().BoolVarP(&f.Neutralize, "neutralize", "N", true, "Neutralize database after restore")
	setCmdIO(cmd, deps)
	return cmd
}
