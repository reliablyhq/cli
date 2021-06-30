package slo

import (
	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reliablyhq/cli/cmd/reliably/alpha/slo/related"
	"github.com/reliablyhq/cli/cmd/reliably/slo/agent"
	init_cmd "github.com/reliablyhq/cli/cmd/reliably/slo/init"
	"github.com/reliablyhq/cli/cmd/reliably/slo/report"
	"github.com/reliablyhq/cli/cmd/reliably/slo/sync"
	"github.com/spf13/cobra"
)

func NewCommand() *cobra.Command {
	cmd := cobra.Command{
		Use:   "slo",
		Short: "service level objective commands",
		Long:  longDescription(),
	}

	cmd.AddCommand(agent.NewCommand(nil))
	cmd.AddCommand(init_cmd.NewCommand(nil))
	cmd.AddCommand(report.NewCommand(nil))
	cmd.AddCommand(sync.NewCommand(nil))
	cmd.AddCommand(related.NewCommand(nil))
	return &cmd
}

func longDescription() string {
	return heredoc.Doc(`A collection of functions to configure and use an SLO`)
}
