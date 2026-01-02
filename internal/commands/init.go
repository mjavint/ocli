package commands

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
)

// initCmd represents the init command
func NewInitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "A brief description of your command",
		Long: `Initialize a new ocli configuration file (ocli.yml) in the current directory.

		This command creates a default configuration file that you can customize
		according to your project needs.`,
		Run: func(cmd *cobra.Command, args []string) {
			configFile := "ocli.yml"

			// Check if config file already exists
			if _, err := os.Stat(configFile); err == nil {
				log.Fatalf("Config file %s already exists", configFile)
			}

			// Default configuration content
			defaultConfig := `odoo:
  config_file: /workspace/odoo.conf
  odoo_bin: /workspace/odoo/odoo-bin
  addons:
    - /workspace/odoo/addons
    - /workspace/enterprise
    - /workspace/custom-addons

db:
  dump_path: /workspace/dbs
  dump_format: zip`
			// Write config file
			if err := os.WriteFile(configFile, []byte(defaultConfig), 0644); err != nil {
				log.Fatalf("failed to create config file: %v", err)
			}

			fmt.Printf("Successfully created %s\n", configFile)
		},
	}
	return cmd
}
