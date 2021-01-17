package doc

import (
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

func NewMarkdownCommand(rootCmd *cobra.Command) *cobra.Command {
	var outputDirectory string = "."

	cmd := &cobra.Command{
		Use:   "markdown",
		Args:  cobra.ExactArgs(0),
		Short: "Generates Markdown pages for the reliably CLI.",
		Long:  "Generate one Markdown document per command.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return doc.GenMarkdownTree(rootCmd, outputDirectory)
		},
	}

	cmd.Flags().StringVar(
		&outputDirectory,
		"output-dir",
		"",
		"Output directory of the generate Markdown documents",
	)

	return cmd
}
