package commands

import (
	"fmt"

	"github.com/mjavint/ocli/internal/odoo"
	"github.com/mjavint/ocli/pkg/utils"
	"github.com/spf13/cobra"
)

// NewRestoredbCmd creates the restoredb command.
func NewRestoredbCmd(deps Deps) *cobra.Command {
	var f odooDBFlags
	cmd := &cobra.Command{
		Use:   "restoredb",
		Short: "Restore an Odoo database from a backup file",
		Long: `Restore an Odoo database from a backup file.
This command executes the Odoo database restore operation using the specified
backup file and Odoo configuration.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			fmt.Fprintln(out, "restoredb called")
			cfg := deps.cfg()
			if f.Bin == "" {
				f.Bin = cfg.Odoo.OdooBin
			}
			if f.Config == "" {
				f.Config = cfg.Odoo.ConfigFile
			}
			if f.NewDB == "" {
				f.NewDB = f.Database
			}
			if f.DumpPath == "" {
				f.DumpPath = cfg.DB.DumpPath
			}

			backupFile := utils.GetBackupFilePath(f.DumpPath, f.Database, cfg.DB.DumpFormat, false)
			fmt.Fprintf(out, "Restoring database: %s from backup file: %s\n", f.NewDB, backupFile)

			f.ForceSet = cmd.Flags().Changed("force")
			f.NeutralizeSet = cmd.Flags().Changed("neutralize")
			cmdArgs := buildRestoreArgs(f, backupFile)
			fmt.Fprintf(out, "Executing: %s %v\n", f.Bin, cmdArgs)

			runner := deps.Runner
			if runner == nil {
				runner = odoo.NewExecRunner()
			}
			if err := runner.Run(cmd.Context(), f.Bin, cmdArgs, odoo.RunOptions{Stdout: out, Stderr: cmd.ErrOrStderr()}); err != nil {
				return fmt.Errorf("Restore failed: %v", err)
			}
			fmt.Fprintf(out, "Restore completed successfully: %s\n", backupFile)
			return nil
		},
	}
	cmd.Flags().StringVarP(&f.Bin, "bin", "b", "", "Path to the Odoo binary")
	cmd.Flags().StringVarP(&f.Config, "config", "c", "", "Odoo configuration file path (odoo.conf)")
	cmd.Flags().StringVarP(&f.NewDB, "new-db", "n", "", "New database name to restore to")
	cmd.Flags().StringVarP(&f.Database, "database", "d", "", "Database name to backup")
	cmd.Flags().StringVarP(&f.DumpPath, "dump-path", "D", "", "Directory to store the backup")
	cmd.Flags().BoolVarP(&f.Force, "force", "f", true, "Force restore even if the database already exists")
	cmd.Flags().BoolVarP(&f.Neutralize, "neutralize", "N", true, "Neutralize database after restore")
	setCmdIO(cmd, deps)
	return cmd
}
