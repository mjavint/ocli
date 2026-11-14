package commands

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// NewInitCmd creates the init command.
func NewInitCmd(deps Deps) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "A brief description of your command",
		Long: `Initialize a new ocli configuration file (ocli.yml) in the current directory.

		This command creates a default configuration file that you can customize
		according to your project needs.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			configFile := "ocli.yml"

			if _, err := os.Stat(configFile); err == nil {
				return fmt.Errorf("Config file %s already exists", configFile)
			}

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
			if err := os.WriteFile(configFile, []byte(defaultConfig), 0644); err != nil {
				return fmt.Errorf("failed to create config file: %v", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Successfully created %s\n", configFile)
			return nil
		},
	}
	setCmdIO(cmd, deps)
	return cmd
}
