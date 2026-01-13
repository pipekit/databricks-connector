package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "databricks-connector",
	Short: "A Databricks connector for Argo Workflows",
	Long:  `A CLI tool to submit, monitor, and retrieve outputs from Databricks jobs, designed to be used as an Argo Workflow Template.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	// Global flags can be defined here, but Databricks SDK often uses env vars.
	// We'll trust the SDK's default config loading (DATABRICKS_HOST, DATABRICKS_TOKEN) for now
	// or add specific flags if requested later.
}
