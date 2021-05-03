package slo

import (
	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reliablyhq/cli/cmd/reliably/slo/apply"
	"github.com/reliablyhq/cli/cmd/reliably/slo/edit"
	init_cmd "github.com/reliablyhq/cli/cmd/reliably/slo/init"
	"github.com/reliablyhq/cli/cmd/reliably/slo/pull"
	"github.com/reliablyhq/cli/cmd/reliably/slo/report"
	"github.com/spf13/cobra"
)

func NewCommand() *cobra.Command {
	cmd := cobra.Command{
		Use:   "slo",
		Short: "service level objective commands",
		Long:  longDescription(),
	}

	cmd.AddCommand(init_cmd.NewCommand())
	cmd.AddCommand(report.NewCommand())
	cmd.AddCommand(pull.NewCommand())
	cmd.AddCommand(apply.NewCommand())
	cmd.AddCommand(edit.NewCommand())

	return &cmd
}

func longDescription() string {
	return heredoc.Doc(`A collection of functions to configure and use an SLO`)
}
