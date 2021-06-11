package sloAlpha

import (
	"github.com/MakeNowJust/heredoc/v2"
	"github.com/spf13/cobra"

	agentAlpha "github.com/reliablyhq/cli/cmd/reliably/alpha/slo/agent"
	initAlpha "github.com/reliablyhq/cli/cmd/reliably/alpha/slo/init"
	initCmd "github.com/reliablyhq/cli/cmd/reliably/slo/init"
)

func NewCommand() *cobra.Command {
	cmd := cobra.Command{
		Use:   "slo",
		Short: "service level objective commands",
		Long:  longDescription(),
	}

	cmd.AddCommand(initCmd.NewCommand(initAlpha.AlpaInitRun))
	cmd.AddCommand(agentAlpha.NewCommand(nil))
	// cmd.AddCommand(report.NewCommand(reportAlpha.AlpaReportRun))
	// cmd.AddCommand(report.NewAlphaCommand(nil))

	return &cmd
}

func longDescription() string {
	return heredoc.Doc(`A collection of functions to configure and use an SLO`)
}
