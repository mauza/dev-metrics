package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "devmetrics",
	Short: "Dev Metrics - A tool to demonstrate flaws in git metrics",
}

var configFile string

func init() {
	rootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", "config.yaml", "Path to config file")

	rootCmd.AddCommand(generateCmd)
	rootCmd.AddCommand(validateCmd)
	rootCmd.AddCommand(setupCmd)
}

// Execute executes the root command
func Execute() error {
	return rootCmd.Execute()
}
