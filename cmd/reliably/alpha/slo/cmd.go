package sloAlpha

import (
	"github.com/MakeNowJust/heredoc/v2"
	"github.com/spf13/cobra"

	//init_cmd "github.com/reliablyhq/cli/cmd/reliably/slo/init"
	"github.com/reliablyhq/cli/cmd/reliably/alpha/slo/report"
	"github.com/reliablyhq/cli/cmd/reliably/slo/report"
)

func NewCommand() *cobra.Command {
	cmd := cobra.Command{
		Use:   "slo",
		Short: "service level objective commands",
		Long:  longDescription(),
	}

	//cmd.AddCommand(init_cmd.NewCommand())
	cmd.AddCommand(report.NewCommand(reportAlpha.AlpaReportRun))
	return &cmd
}

func longDescription() string {
	return heredoc.Doc(`A collection of functions to configure and use an SLO`)
}
