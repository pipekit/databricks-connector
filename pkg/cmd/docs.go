package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

var docsCmd = &cobra.Command{
	Use:   "docs",
	Short: "Generate documentation for the CLI",
	Long:  `Generate Markdown documentation for the CLI commands in the docs/ directory.`,
	Hidden: true, // Hide this command from the main help output if preferred
	RunE: func(cmd *cobra.Command, args []string) error {
		dir := "./docs"
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			if err := os.Mkdir(dir, 0755); err != nil {
				return err
			}
		}

		fmt.Printf("Generating documentation in %s...\n", dir)
		// GenMarkdownTree generates markdown for the command and all its children
		return doc.GenMarkdownTree(rootCmd, dir)
	},
}

func init() {
	rootCmd.AddCommand(docsCmd)
}

