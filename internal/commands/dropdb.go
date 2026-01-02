package commands

import (
	"fmt"
	"log"
	"os/exec"

	"github.com/mjavint/ocli/pkg/config"
	"github.com/spf13/cobra"
)

// dropdbCmd represents the dropdb command
func NewDropdbCmd() *cobra.Command {
	var (
		odooBin    string
		configPath string
		dbName     string
	)

	cmd := &cobra.Command{
		Use:   "dropdb",
		Short: "A brief description of your command",
		Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("dropdb called")
			//odoo-bin db -c /etc/odoo.conf drop db_name
			if odooBin == "" {
				odooBin = config.AppConfig.Odoo.OdooBin
			}
			if configPath == "" {
				configPath = config.AppConfig.Odoo.ConfigFile
			}
			if dbName == "" {
				log.Fatal("Database name is required. Use --database or -d to specify it.")
			}

			// Execute odoo-bin db drop command
			cmdExec := exec.Command(odooBin, "db", "-c", configPath, "drop", dbName)

			// Capture output
			output, err := cmdExec.CombinedOutput()
			if err != nil {
				log.Fatalf("Error executing drop command: %v\nOutput: %s", err, string(output))
			}

			fmt.Printf("Database drop completed successfully: %s\n", dbName)
		},
	}
	cmd.Flags().StringVarP(&odooBin, "bin", "b", "", "Path to the Odoo binary")
	cmd.Flags().StringVarP(&configPath, "config", "c", "", "Odoo configuration file path (odoo.conf)")
	cmd.Flags().StringVarP(&dbName, "database", "d", "", "Database name to backup")
	return cmd

}
