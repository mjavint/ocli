package commands

import (
	"fmt"
	"log"
	"os/exec"

	"github.com/mjavint/ocli/pkg/config"
	"github.com/spf13/cobra"
)

// initdbCmd represents the initdb command
func NewInitDBCmd() *cobra.Command {
	var (
		odooBin    string
		configPath string
		dbName     string
		// Opcionales
		withDemo bool
		force    bool
		lang     string
		username string
		password string
		country  string
	)
	cmd := &cobra.Command{
		Use:   "initdb",
		Short: "A brief description of your command",
		Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("initdb called")
			//odoo-bin db -c /etc/odoo.conf init db_name
			if odooBin == "" {
				odooBin = config.AppConfig.Odoo.OdooBin
			}
			if configPath == "" {
				configPath = config.AppConfig.Odoo.ConfigFile
			}
			if dbName == "" {
				log.Fatal("Database name is required. Use --database or -d to specify it.")
			}

			// Build command arguments dynamically based on optional flags
			cmdArgs := []string{"db", "-c", configPath, "init", dbName}
			// Adjust withDemo flag - if user provides --without-demo, set withDemo to true
			if cmd.Flags().Changed("with-demo") {
				withDemo = !withDemo
			}
			if !withDemo {
				cmdArgs = append(cmdArgs, "--with-demo")
			}

			if cmd.Flags().Changed("force") {
				force = !force
			}
			if !force {
				cmdArgs = append(cmdArgs, "--force")
			}
			if lang != "" {
				cmdArgs = append(cmdArgs, "--language="+lang)
			}
			if username != "" {
				cmdArgs = append(cmdArgs, "--username="+username)
			}
			if password != "" {
				cmdArgs = append(cmdArgs, "--password="+password)
			}
			if country != "" {
				cmdArgs = append(cmdArgs, "--country="+country)
			}

			// Execute odoo-bin db init command with dynamically built arguments
			fmt.Printf("Executing command: %s %v\n", odooBin, cmdArgs)
			cmdExec := exec.Command(odooBin, cmdArgs...)
			// Capture output
			output, err := cmdExec.CombinedOutput()
			if err != nil {
				log.Fatalf("Error executing init command: %v\nOutput: %s", err, string(output))
			}

			fmt.Printf("Database init completed successfully: %s\n", dbName)
		},
	}
	cmd.Flags().StringVarP(&odooBin, "bin", "b", "", "Path to the Odoo binary")
	cmd.Flags().StringVarP(&configPath, "config", "c", "", "Odoo configuration file path (odoo.conf)")
	cmd.Flags().StringVarP(&dbName, "database", "d", "", "Database name to backup")
	cmd.Flags().BoolVar(&withDemo, "with-demo", true, "Load demo data")
	cmd.Flags().BoolVar(&force, "force", true, "Force database creation if it already exists")
	cmd.Flags().StringVar(&lang, "lang", "", "Language code for the new database")
	cmd.Flags().StringVar(&username, "username", "", "Administrator username")
	cmd.Flags().StringVar(&password, "password", "", "Administrator password")
	cmd.Flags().StringVar(&country, "country", "", "Country code for localization")
	return cmd
}
