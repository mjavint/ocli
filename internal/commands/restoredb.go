package commands

import (
	"fmt"
	"log"
	"os/exec"

	"github.com/mjavint/ocli/pkg/config"
	"github.com/mjavint/ocli/pkg/utils"
	"github.com/spf13/cobra"
)

// restoredbCmd represents the restoredb command
func NewRestoredbCmd() *cobra.Command {
	var (
		odooBin    string
		configPath string
		dbName     string
		newName    string
		backupDir  string
		force      bool
		neutralize bool
	)
	cmd := &cobra.Command{
		Use:   "restoredb",
		Short: "Restore an Odoo database from a backup file",
		Long: `Restore an Odoo database from a backup file.
This command executes the Odoo database restore operation using the specified
backup file and Odoo configuration.`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("restoredb called")
			// odoo-bin db -c config.conf load new_db /path/to/backup.zip
			if odooBin == "" {
				odooBin = config.AppConfig.Odoo.OdooBin
			}
			if configPath == "" {
				configPath = config.AppConfig.Odoo.ConfigFile
			}

			if newName == "" {
				newName = dbName
			}
			if backupDir == "" {
				backupDir = config.AppConfig.DB.DumpPath
			}

			backupFile := utils.GetBackupFilePath(backupDir, dbName, config.AppConfig.DB.DumpFormat, false)

			fmt.Printf("Restoring database: %s from backup file: %s\n", newName, backupFile)

			// Build command arguments: odoo-bin db -c config load new_db backup_file
			cmdArgs := []string{"db", "-c", configPath, "load", newName, backupFile}

			if cmd.Flags().Changed("force") {
				force = !force
			}

			if cmd.Flags().Changed("neutralize") {
				neutralize = !neutralize
			}

			if !force {
				cmdArgs = append(cmdArgs, "--force")
			}
			if !neutralize {
				cmdArgs = append(cmdArgs, "--neutralize")
			}

			fmt.Printf("Executing: %s %v\n", odooBin, cmdArgs)

			// Execute odoo-bin db dump command
			cmdExec := exec.Command(odooBin, cmdArgs...)

			// Run the command
			if err := cmdExec.Run(); err != nil {
				log.Fatalf("Restore failed: %v", err)
			}

			fmt.Printf("Restore completed successfully: %s\n", backupFile)
		},
	}
	cmd.Flags().StringVarP(&odooBin, "bin", "b", "", "Path to the Odoo binary")
	cmd.Flags().StringVarP(&configPath, "config", "c", "", "Odoo configuration file path (odoo.conf)")
	cmd.Flags().StringVarP(&newName, "new-db", "n", "", "New database name to restore to")
	cmd.Flags().StringVarP(&dbName, "database", "d", "", "Database name to backup")
	cmd.Flags().StringVarP(&backupDir, "dump-path", "D", "", "Directory to store the backup")
	cmd.Flags().BoolVarP(&force, "force", "f", true, "Force restore even if the database already exists")
	cmd.Flags().BoolVarP(&neutralize, "neutralize", "N", true, "Neutralize database after restore")
	return cmd
}
