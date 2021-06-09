package agent

import (
	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reliablyhq/cli/agent"
	"github.com/reliablyhq/cli/core/iostreams"
	"github.com/reliablyhq/cli/core/manifest"
	"github.com/reliablyhq/cli/core/metrics"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type Options struct {
	IO           *iostreams.IOStreams
	ManifestPath string
	Interval     int64
}

var longDesc = heredoc.Doc(`
	Runs cli in SLO agent mode. This mode utilises data defined
	in the slo manifest to retrieve metrics and generate indicators.

	The indicators are sent to reliably.`)

var examples = heredoc.Doc(`
	$ reliably alpha slo agent -m reliably.yaml -i 300
	$ reliably alpha slo agent --interval 600`)

func NewCommand(runF func(*Options) error) *cobra.Command {
	opts := &Options{
		IO: iostreams.System(),
	}

	cmd := &cobra.Command{
		Use:     "agent",
		Short:   "runs cli in agent mode sending SLOs to reliably",
		Long:    longDesc,
		Example: examples,
		RunE: func(cmd *cobra.Command, args []string) error {
			logger := log.StandardLogger()
			var objectives []*agent.JobObjective
			job := agent.NewJob(opts.Interval, objectives, metrics.GCPProvider)

			job.ErrorFunc(func(e *agent.Error) {
				logger.Errorf(
					"error processing objective: %s\nerror: %s",
					e.Objective, e.Error())
			}).Do()

			return nil
		},
	}

	// define flags
	cmd.Flags().StringVarP(&opts.ManifestPath, "manifest", "m", manifest.DefaultManifestPath, "the location of the manifest file")
	cmd.Flags().Int64VarP(&opts.Interval, "interval", "i", 300, "interval indicators are pushed at in seconds")
	return cmd
}
