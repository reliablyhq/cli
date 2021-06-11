package sloAlpha

import (
	"github.com/MakeNowJust/heredoc/v2"
	"github.com/spf13/cobra"

	agentAlpha "github.com/reliablyhq/cli/cmd/reliably/alpha/slo/agent"
	initAlpha "github.com/reliablyhq/cli/cmd/reliably/alpha/slo/init"
	reportAlpha "github.com/reliablyhq/cli/cmd/reliably/alpha/slo/report"
	"github.com/reliablyhq/cli/cmd/reliably/alpha/slo/sync"
	initCmd "github.com/reliablyhq/cli/cmd/reliably/slo/init"
	"github.com/reliablyhq/cli/cmd/reliably/slo/report"
)

func NewCommand() *cobra.Command {
	cmd := cobra.Command{
		Use:   "slo",
		Short: "service level objective commands",
		Long:  longDescription(),
	}

	cmd.AddCommand(initCmd.NewCommand(initAlpha.AlpaInitRun))
	cmd.AddCommand(agentAlpha.NewCommand(nil))
	cmd.AddCommand(report.NewCommand(reportAlpha.AlpaReportRun))
	cmd.AddCommand(sync.NewCommand(nil))

	return &cmd
}

func longDescription() string {
	return heredoc.Doc(`A collection of functions to configure and use an SLO`)
}
