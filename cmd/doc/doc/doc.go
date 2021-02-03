package doc

import (
	"github.com/spf13/cobra"
)

func NewCmdDoc(rootCmd *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:    "",
		Short:  "Generate reliably CLI documentation",
		Long:   `Generate the documentation of all the reliably CLI commands.`,
		Hidden: true,
	}

	cmd.AddCommand(NewManCommand(rootCmd))
	cmd.AddCommand(NewMarkdownCommand(rootCmd))

	return cmd
}
