package internal

import (
	"os"

	"github.com/mjavint/ocli/internal/commands"
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "ocli",
	Short: "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.ocli.yaml)")
	// Cobra also supports local flags, which will only run
	rootCmd.AddCommand(commands.NewInitCmd())
	rootCmd.AddCommand(commands.NewListdbCmd())
	rootCmd.AddCommand(commands.NewBackupdbCmd())
	rootCmd.AddCommand(commands.NewRestoredbCmd())
	rootCmd.AddCommand(commands.NewCopydbCmd())
	rootCmd.AddCommand(commands.NewInitDBCmd())
	rootCmd.AddCommand(commands.NewDropdbCmd())
	rootCmd.AddCommand(commands.NewRenamedbCmd())
	rootCmd.AddCommand(commands.NewConfigAddonCmd())
	rootCmd.AddCommand(commands.NewStartOdooCmd())
}
