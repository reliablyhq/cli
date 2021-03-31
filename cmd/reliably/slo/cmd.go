package slo

import (
	init_cmd "github.com/reliablyhq/cli/cmd/reliably/slo/init"
	"github.com/reliablyhq/cli/cmd/reliably/slo/report"
	"github.com/spf13/cobra"
)

func NewCommand() *cobra.Command {
	cmd := cobra.Command{
		Use:   "slo",
		Short: "service level objective",
		Long:  longDescription(),
	}

	cmd.AddCommand(init_cmd.NewCommand())
	cmd.AddCommand(report.NewCommand())

	return &cmd
}

func longDescription() string {
	return "todo: write something here"
}
