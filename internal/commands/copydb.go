package commands

import (
	"fmt"
	"log"
	"os/exec"

	"github.com/mjavint/ocli/pkg/config"
	"github.com/spf13/cobra"
)

// copydbCmd represents the copydb command
func NewCopydbCmd() *cobra.Command {
	var (
		odooBin    string
		configPath string
		dbName     string
		newName    string
		force      bool
		neutralize bool
	)
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
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("copydb called")
			//odoo-bin db -c config.conf duplicate db_name new_dbname
			if odooBin == "" {
				odooBin = config.AppConfig.Odoo.OdooBin
			}
			if configPath == "" {
				configPath = config.AppConfig.Odoo.ConfigFile
			}
			if dbName == "" {
				log.Fatal("Database name is required. Use --database or -d to specify it.")
			}
			if newName == "" {
				log.Fatal("New database name is required. Use --new-db or -n to specify it.")
			}

			// Execute odoo-bin db duplicate command
			cmdArgs := []string{"db", "-c", configPath, "duplicate", dbName, newName}

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

			cmdExec := exec.Command(odooBin, cmdArgs...)

			// Run the command
			if err := cmdExec.Run(); err != nil {
				log.Fatalf("Error executing duplicate command: %v", err)
			}

			fmt.Printf("Duplicate completed successfully: %s\n", newName)
		},
	}
	cmd.Flags().StringVarP(&odooBin, "bin", "b", "", "Path to the Odoo binary")
	cmd.Flags().StringVarP(&configPath, "config", "c", "", "Odoo configuration file path (odoo.conf)")
	cmd.Flags().StringVarP(&dbName, "database", "d", "", "Database name to backup")
	cmd.Flags().StringVarP(&newName, "new-db", "n", "", "New database name to restore to")
	cmd.Flags().BoolVarP(&force, "force", "f", true, "Force restore even if the database already exists")
	cmd.Flags().BoolVarP(&neutralize, "neutralize", "N", true, "Neutralize database after restore")
	return cmd
}
