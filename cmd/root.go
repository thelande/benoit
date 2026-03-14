/*
Copyright © 2026 Tom Helander <thomas.helander@gmail.com>
*/
package cmd

import (
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "benoit",
	Short: "An automated diagnostic engine.",
	Long: `Benoit is a user-defined automated diagnostic tool.

	It runs commands and, depending on the exit code, runs additional commands
	if the command succeeded or failed. This allows for chains of diagnostics
	to be built, with additional steps such as log collection to be performed
	if/when a check fails.
	`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		return initializeConfig(cmd)
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {}

func initializeConfig(cmd *cobra.Command) error {
	viper.SetEnvPrefix("CDE")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "*", "-", "*"))
	viper.AutomaticEnv()

	err := viper.BindPFlags(cmd.Flags())
	if err != nil {
		return err
	}

	return nil
}
