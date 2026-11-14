package commands

import (
	"fmt"

	"github.com/mjavint/ocli/pkg/config"
	"github.com/mjavint/ocli/pkg/db"
	"github.com/spf13/cobra"
)

// NewListdbCmd creates the listdb command.
func NewListdbCmd(deps Deps) *cobra.Command {
	var odooConfigFile string
	cmd := &cobra.Command{
		Use:   "listdb",
		Short: "A brief description of your command",
		Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			fmt.Fprintln(out, "listdb called")

			lister := deps.ListDatabases
			if lister == nil {
				lister = db.ListDatabasesWithSizesConfig
			}

			if odooConfigFile == "" {
				odooConfigFile = deps.cfg().Odoo.ConfigFile
			}

			var pgCfg *db.PGConfig
			if deps.ListDatabases == nil {
				dbConfig, err := config.LoadOdooDBParams(odooConfigFile)
				if err != nil {
					return fmt.Errorf("Error resolviendo configuración: %v", err)
				}
				pgCfg = &db.PGConfig{
					Host:     dbConfig.Host,
					Port:     dbConfig.Port,
					User:     dbConfig.User,
					Password: dbConfig.Password,
				}
			} else {
				pgCfg = &db.PGConfig{}
			}

			databases, err := lister(cmd.Context(), pgCfg)
			if err != nil {
				return fmt.Errorf("Error listing databases: %v", err)
			}

			fmt.Fprintf(out, "\nFound %d database(s):\n\n", len(databases))
			for _, d := range databases {
				fmt.Fprintf(out, "- %s (%s)\n", d.Name, d.Size)
			}
			fmt.Fprintln(out)
			return nil
		},
	}
	cmd.Flags().StringVarP(&odooConfigFile, "config", "c", "", "Odoo configuration file path (odoo.conf)")
	setCmdIO(cmd, deps)
	return cmd
}
