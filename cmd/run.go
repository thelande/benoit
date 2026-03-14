/*
Copyright © 2026 Tom Helander <thomas.helander@gmail.com>
*/
package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	benoit "github.com/thelande/benoit/pkg/benoit"
	"go.yaml.in/yaml/v2"
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:          "run",
	Short:        "Run a pipeline",
	Long:         `Load an execution pipeline from a file and run it.`,
	SilenceUsage: true,
	RunE:         run,
}

func init() {
	rootCmd.AddCommand(runCmd)
	runCmd.Flags().StringP("input", "i", "", "Input file path")
}

// Load the input file, create the execution engine, and run it.
func run(cmd *cobra.Command, args []string) error {
	var spec benoit.Spec
	data, err := os.ReadFile(viper.GetString("input"))
	if err != nil {
		return err
	}

	if err := yaml.Unmarshal(data, &spec); err != nil {
		panic(err)
	}

	engine, err := benoit.NewEngine(spec)
	if err != nil {
		return err
	}

	err = engine.Run(spec.Metadata.Start)
	if err != nil {
		return err
	}

	return nil
}
