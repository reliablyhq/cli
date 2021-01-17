package cmd

import (
	"github.com/spf13/cobra"

	docCmd "github.com/reliablyhq/cli/cmd/doc"
)

func NewCmdDoc() *cobra.Command {
	cmd := &cobra.Command{
		Use:    "doc <command>",
		Short:  "Generate reliably CLI documentation",
		Long:   `Generate the documentation of all the reliably CLI commands.`,
		Hidden: true,
	}

	cmd.AddCommand(docCmd.NewManCommand(rootCmd))
	cmd.AddCommand(docCmd.NewMarkdownCommand(rootCmd))

	return cmd
}

func init() {
	rootCmd.AddCommand(NewCmdDoc())
}
