package commands

import (
	"fmt"
	"log"
	"os/exec"

	"github.com/mjavint/ocli/pkg/config"
	"github.com/mjavint/ocli/pkg/utils"
	"github.com/spf13/cobra"
)

// backupdbCmd represents the backupdb command
func NewBackupdbCmd() *cobra.Command {
	var (
		odooBin      string
		configPath   string
		dbName       string
		dumpPath     string
		backupFormat string
		noFilestore  bool
	)
	cmd := &cobra.Command{
		Use:   "backupdb",
		Short: "Backup an Odoo database",
		Long: `Backup an Odoo database using the Odoo binary.
This command executes the Odoo database backup operation and stores
the backup file in the specified directory.`,
		Run: func(cmd *cobra.Command, args []string) {
			// Validate required flags
			if odooBin == "" {
				odooBin = config.AppConfig.Odoo.OdooBin
			}
			if configPath == "" {
				configPath = config.AppConfig.Odoo.ConfigFile
			}
			if dbName == "" {
				log.Fatal("Database name is required. Use --database or -d to specify it.")
			}
			if dumpPath == "" {
				dumpPath = config.AppConfig.DB.DumpPath
			}
			if backupFormat == "" {
				backupFormat = config.AppConfig.DB.DumpFormat
			}

			fmt.Printf("Backing up database: %s\n", dbName)

			// Build the dump file path first
			dumpFile := utils.GetBackupFilePath(dumpPath, dbName, backupFormat, noFilestore)
			// Build command arguments: odoo-bin db -c config dump database output_file -f format
			cmdArgs := []string{"db", "-c", configPath, "dump", dbName, dumpFile}

			// Add no-filestore flag if specified
			if noFilestore {
				cmdArgs = append(cmdArgs, "--no-filestore")
			}

			fmt.Printf("Executing: %s %v\n", odooBin, cmdArgs)

			// Execute odoo-bin db dump command
			cmdExec := exec.Command(odooBin, cmdArgs...)

			// Run the command
			if err := cmdExec.Run(); err != nil {
				log.Fatalf("Error executing backup command: %v", err)
			}

			fmt.Printf("Backup completed successfully: %s\n", dumpFile)
		},
	}
	cmd.Flags().StringVarP(&odooBin, "bin", "b", "", "Path to the Odoo binary")
	cmd.Flags().StringVarP(&configPath, "config", "c", "", "Odoo configuration file path (odoo.conf)")
	cmd.Flags().StringVarP(&dbName, "database", "d", "", "Database name to backup")
	cmd.Flags().StringVarP(&dumpPath, "dump-path", "D", "", "Directory to store the backup")
	cmd.Flags().StringVarP(&backupFormat, "format", "f", "", "Backup file format (e.g., zip, tar)")
	cmd.Flags().BoolVar(&noFilestore, "no-filestore", false, "Exclude filestore from backup")

	return cmd
}
