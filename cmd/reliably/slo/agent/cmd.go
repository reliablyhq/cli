package agent

import (
	"fmt"

	"github.com/MakeNowJust/heredoc/v2"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/reliablyhq/cli/api"
	"github.com/reliablyhq/cli/config"
	"github.com/reliablyhq/cli/core/agent"
	"github.com/reliablyhq/cli/core/entities"
	"github.com/reliablyhq/cli/core/iostreams"
	"github.com/reliablyhq/cli/core/manifest"
)

type AgentOptions struct {
	IO           *iostreams.IOStreams
	ManifestPath string
	Interval     int64
	Selector     string
	ReportUI     bool
}

var longDesc = heredoc.Doc(`
	Runs the CLI in SLO agent mode. This mode utilises data defined
	in the slo manifest to retrieve metrics and generate indicators.

	The indicators are sent to reliably.`)

var examples = heredoc.Doc(`
	$ reliably slo agent -m reliably.yaml -i 300
	$ reliably slo agent --interval 600`)

func NewCommand(runF func(*AgentOptions) error) *cobra.Command {
	opts := &AgentOptions{
		IO: iostreams.System(),
	}

	cmd := &cobra.Command{
		Use:     "agent",
		Short:   "runs in agent mode sending SLIs to Reliably",
		Long:    longDesc,
		Example: examples,
		RunE: func(cmd *cobra.Command, args []string) error {
			if runF != nil {
				return runF(opts)
			}

			return agentRun(opts)

		},
	}

	// define flags
	cmd.Flags().StringVarP(&opts.ManifestPath, "manifest", "m", manifest.DefaultManifestPath, "the location of the manifest file")
	cmd.Flags().Int64VarP(&opts.Interval, "interval", "i", 300, "interval indicators are pushed at in seconds")
	cmd.Flags().StringVarP(&opts.Selector, "selector", "l", "", "objectives selector based on labels - only used when --report-ui/-R flag is used")
	cmd.Flags().BoolVarP(&opts.ReportUI, "report-view", "R", false, "shows a view of the report while pushing indicators")
	return cmd
}

func agentRun(opts *AgentOptions) error {
	logger := log.StandardLogger()

	var m entities.Manifest
	if err := m.LoadFromFile(opts.ManifestPath); err != nil {
		return err
	}

	fmt.Println(m)

	// get API client
	org, err := config.GetCurrentOrgInfo()
	if err != nil {
		return err
	}

	// define agent job
	client := api.NewClientFromHTTP(api.AuthHTTPClient(config.Hostname))

	if opts.ReportUI {
		return runUI(client, opts, &m, org.Name)
	}

	agent.SetLogger(logger)
	job := agent.NewJob(opts.Interval, m)
	job.ErrorFunc(func(e *agent.Error) {
		logger.Errorf(
			"error processing objective: %v\nerror: %s",
			e.Objective, e.Error())
	}).IndicatorFunc(func(i *entities.Indicator) error {
		return api.CreateEntity(client, config.Hostname, org.Name, i)
	}).Do()

	return nil
}
