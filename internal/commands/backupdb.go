package commands

import (
	"fmt"

	"github.com/mjavint/ocli/internal/odoo"
	"github.com/mjavint/ocli/pkg/utils"
	"github.com/spf13/cobra"
)

// NewBackupdbCmd creates the backupdb command.
func NewBackupdbCmd(deps Deps) *cobra.Command {
	var f odooDBFlags
	cmd := &cobra.Command{
		Use:   "backupdb",
		Short: "Backup an Odoo database",
		Long: `Backup an Odoo database using the Odoo binary.
This command executes the Odoo database backup operation and stores
the backup file in the specified directory.`,
		RunE: func(cmd *cobra.Command, args []string) error {
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
			if f.DumpPath == "" {
				f.DumpPath = cfg.DB.DumpPath
			}
			if f.DumpFormat == "" {
				f.DumpFormat = cfg.DB.DumpFormat
			}

			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "Backing up database: %s\n", f.Database)
			dumpFile, err := utils.GetBackupFilePath(f.DumpPath, f.Database, f.DumpFormat, f.NoFilestore)
			if err != nil {
				return fmt.Errorf("Error preparing backup path: %v", err)
			}
			cmdArgs := buildBackupArgs(f, dumpFile)
			fmt.Fprintf(out, "Executing: %s %v\n", f.Bin, cmdArgs)

			runner := deps.Runner
			if runner == nil {
				runner = odoo.NewExecRunner()
			}
			if err := runner.Run(cmd.Context(), f.Bin, cmdArgs, odoo.RunOptions{Stdout: out, Stderr: cmd.ErrOrStderr()}); err != nil {
				return fmt.Errorf("Error executing backup command: %v", err)
			}
			fmt.Fprintf(out, "Backup completed successfully: %s\n", dumpFile)
			return nil
		},
	}
	cmd.Flags().StringVarP(&f.Bin, "bin", "b", "", "Path to the Odoo binary")
	cmd.Flags().StringVarP(&f.Config, "config", "c", "", "Odoo configuration file path (odoo.conf)")
	cmd.Flags().StringVarP(&f.Database, "database", "d", "", "Database name to backup")
	cmd.Flags().StringVarP(&f.DumpPath, "dump-path", "D", "", "Directory to store the backup")
	cmd.Flags().StringVarP(&f.DumpFormat, "format", "f", "", "Backup file format (e.g., zip, tar)")
	cmd.Flags().BoolVar(&f.NoFilestore, "no-filestore", false, "Exclude filestore from backup")
	setCmdIO(cmd, deps)
	return cmd
}
