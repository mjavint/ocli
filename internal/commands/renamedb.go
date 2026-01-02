package commands

import (
	"fmt"
	"log"
	"os/exec"

	"github.com/mjavint/ocli/pkg/config"
	"github.com/spf13/cobra"
)

// renamedbCmd represents the renamedb command
func NewRenamedbCmd() *cobra.Command {
	var (
		odooBin    string
		configPath string
		dbName     string
		newName    string
		force      bool
	)
	cmd := &cobra.Command{
		Use:   "renamedb",
		Short: "Rename an Odoo database",
		Long: `Rename an Odoo database from one name to another.
This command connects to the Odoo database management API
and performs a database rename operation.

Example:
  ocli renamedb mydb_old mydb_new`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("db called")
			// odoo-bin db -c /etc/odoo.conf rename db_name new_name
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
			cmdArgs := []string{"db", "-c", configPath, "rename", dbName, newName}

			if condition := cmd.Flags().Changed("force"); condition {
				cmdArgs = append(cmdArgs, "--force")
			}

			cmdExec := exec.Command(odooBin, cmdArgs...)

			// Run the command
			if err := cmdExec.Run(); err != nil {
				log.Fatalf("Rename failed: %v", err)
			}

			fmt.Printf("Rename completed successfully: %s\n", newName)
		},
	}
	cmd.Flags().StringVarP(&odooBin, "bin", "b", "", "Path to the Odoo binary")
	cmd.Flags().StringVarP(&configPath, "config", "c", "", "Odoo configuration file path (odoo.conf)")
	cmd.Flags().StringVarP(&newName, "new-db", "n", "", "New database name to restore to")
	cmd.Flags().StringVarP(&dbName, "database", "d", "", "Database name to backup")
	cmd.Flags().BoolVarP(&force, "force", "f", true, "Force restore even if the database already exists")
	return cmd
}
