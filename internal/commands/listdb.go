package commands

import (
	"context"
	"fmt"
	"io"

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
			if odooConfigFile == "" {
				odooConfigFile = deps.cfg().Odoo.ConfigFile
			}

			dbConfig, err := config.LoadOdooDBParams(odooConfigFile)
			if err != nil {
				return fmt.Errorf("Error resolviendo configuración: %v", err)
			}
			pgCfg := &db.PGConfig{
				Host:     dbConfig.Host,
				Port:     dbConfig.Port,
				User:     dbConfig.User,
				Password: dbConfig.Password,
			}
			dblist, err := db.ListDatabases(cmd.Context(), pgCfg)
			if err != nil {
				return fmt.Errorf("Error listing databases: %v", err)
			}
			if err := displayTable(cmd.Context(), dblist, pgCfg, out); err != nil {
				return fmt.Errorf("Error displaying databases: %v", err)
			}
			return nil
		},
	}
	cmd.Flags().StringVarP(&odooConfigFile, "config", "c", "", "Odoo configuration file path (odoo.conf)")
	setCmdIO(cmd, deps)
	return cmd
}

// displayTable shows results in table format.
func displayTable(ctx context.Context, databases []string, pgCfg *db.PGConfig, out io.Writer) error {
	fmt.Fprintf(out, "\nFound %d database(s):\n\n", len(databases))
	for _, dbname := range databases {
		size, err := db.GetDatabaseSize(ctx, dbname, pgCfg)
		if err != nil {
			return fmt.Errorf("error getting size for database %s: %w", dbname, err)
		}
		fmt.Fprintf(out, "- %s (%s)\n", dbname, size)
	}

	fmt.Fprintln(out)
	return nil
}
