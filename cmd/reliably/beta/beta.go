package beta

import (
	"github.com/MakeNowJust/heredoc/v2"
	"github.com/spf13/cobra"
)

func NewCommand() *cobra.Command {
	cmd := cobra.Command{
		Use:    "beta",
		Short:  "Beta versions of reliably commands",
		Long:   longDescription(),
		Hidden: true,
	}

	//cmd.AddCommand()

	return &cmd
}

func longDescription() string {
	return heredoc.Doc(`A set of commands that are in their beta version of development.`)
}
