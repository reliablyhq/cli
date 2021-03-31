package slo

import (
	"github.com/reliablyhq/cli/cmd/reliably/slo/report"
	"github.com/spf13/cobra"
)

func NewCommand() *cobra.Command {
	cmd := cobra.Command{
		Use:   "slo",
		Short: "service level objective",
		Long:  longDescription(),
	}

	cmd.AddCommand(report.NewCommand())

	return &cmd
}

func longDescription() string {
	return "todo: write something here"
}
