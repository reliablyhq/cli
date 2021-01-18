package doc

import (
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

func NewManCommand(rootCmd *cobra.Command) *cobra.Command {
	var outputDirectory string = "."

	cmd := &cobra.Command{
		Use:   "man",
		Args:  cobra.ExactArgs(0),
		Short: "Generates man pages for the reliably CLI.",
		Long:  "Generate one man page per command.",
		RunE: func(cmd *cobra.Command, args []string) error {
			header := doc.GenManHeader{
				Section: "1",
				Manual:  "Reliably Manual",
			}
			return doc.GenManTree(rootCmd, &header, outputDirectory)
		},
	}

	cmd.Flags().StringVar(
		&outputDirectory,
		"output-dir",
		"",
		"Output directory of the generate man pages",
	)

	return cmd
}
