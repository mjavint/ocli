package commands

import (
	"context"
	"fmt"
	"log"

	"github.com/mjavint/ocli/pkg/config"
	"github.com/mjavint/ocli/pkg/db"
	"github.com/spf13/cobra"
)

// listdbCmd represents the listdb command
func NewListdbCmd() *cobra.Command {
	var odooConfigFile string
	cmd := &cobra.Command{
		Use:   "listdb",
		Short: "A brief description of your command",
		Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("listdb called")
			if odooConfigFile == "" {
				odooConfigFile = config.AppConfig.Odoo.ConfigFile
			}

			dbConfig, err := config.LoadOdooDBParams(odooConfigFile)
			if err != nil {
				log.Fatalf("Error resolviendo configuración: %v", err)
			}
			// Construir configuración de PostgreSQL
			pgCfg := &db.PGConfig{
				Host:     dbConfig.Host,
				Port:     dbConfig.Port,
				User:     dbConfig.User,
				Password: dbConfig.Password,
			}
			// Listar bases de datos
			dblist, err := db.ListDatabases(cmd.Context(), pgCfg)
			if err != nil {
				log.Fatalf("Error listing databases: %v", err)
			}
			if err := displayTable(cmd.Context(), dblist, pgCfg); err != nil {
				log.Fatalf("Error displaying databases: %v", err)
			}
		},
	}
	// Definir flags
	cmd.Flags().StringVarP(&odooConfigFile, "config", "c", "", "Odoo configuration file path (odoo.conf)")
	return cmd
}

// displayTable shows results in table format
func displayTable(ctx context.Context, databases []string, pgCfg *db.PGConfig) error {
	fmt.Printf("\nFound %d database(s):\n\n", len(databases))
	for _, dbname := range databases {
		size, err := db.GetDatabaseSize(ctx, dbname, pgCfg)
		if err != nil {
			return fmt.Errorf("error getting size for database %s: %w", dbname, err)
		}
		fmt.Printf("- %s (%s)\n", dbname, size)
	}

	fmt.Println()
	return nil
}
