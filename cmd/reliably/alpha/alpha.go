package alpha

import (
	"github.com/MakeNowJust/heredoc/v2"
	"github.com/spf13/cobra"
)

func NewCommand() *cobra.Command {
	cmd := cobra.Command{
		Use:    "alpha",
		Short:  "Alpha versions of reliably commands",
		Long:   longDescription(),
		Hidden: true,
	}

	return &cmd
}

func longDescription() string {
	return heredoc.Doc(`A set of commands that are in their alpha version of development.`)
}
