package history

import (
	//"errors"
	"fmt"
	"net/http"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/reliablyhq/cli/api"
	"github.com/reliablyhq/cli/cmd/reliably/cmdutil"
	"github.com/reliablyhq/cli/core"
	"github.com/reliablyhq/cli/core/iostreams"
	"github.com/reliablyhq/cli/core/report"
)

type HistoryOptions struct {
	IO         *iostreams.IOStreams
	HttpClient func() *http.Client
	Hostname   string

	OrgID string

	LimitResults int
}

func NewCmdHistory() *cobra.Command {
	opts := &HistoryOptions{
		IO: iostreams.System(),
		HttpClient: func() *http.Client {
			return api.AuthHTTPClient(core.Hostname())
		},
	}

	cmd := &cobra.Command{
		Use:   "history",
		Short: "Show your SLO history",
		Long:  `Show the evolution of your SLOs over time.`,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			if !cmdutil.CheckAuth() {
				cmdutil.PrintRequireAuthMsg()
				os.Exit(1)
			}
		},
		PreRunE: func(cmd *cobra.Command, args []string) error {
			var err error
			apiClient := api.NewClientFromHTTP(opts.HttpClient())

			// Ensure the CLI history is executed in a valid org
			opts.OrgID, err = api.CurrentUserOrganizationID(apiClient, core.Hostname())
			if err != nil {
				return fmt.Errorf("unable to retrieve current organization: %w", err)
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Hostname = core.Hostname()

			if opts.LimitResults < 1 {
				return fmt.Errorf("invalid value for --limit: %v", opts.LimitResults)
			}

			return historyRun(opts)
		},
	}

	cmd.Flags().IntVarP(&opts.LimitResults, "limit", "l", 10, "Maximum number of reports to fetch")

	return cmd
}

func historyRun(opts *HistoryOptions) (err error) {

	//opts.IO.StartProgressIndicator()

	apiClient := api.NewClientFromHTTP(opts.HttpClient())
	reports, err := api.GetReports(apiClient, opts.Hostname, opts.OrgID, opts.LimitResults)

	//opts.IO.StopProgressIndicator()

	if err != nil {
		return fmt.Errorf("Unable to retrieve your history of SLO reports: %w", err)
	}

	fmt.Printf("We fetched %d reports\n", len(reports))
	for i, r := range reports {
		fmt.Printf("Report #%d\n", i+1)
		report.Write(report.SimpleText, &r, opts.IO.Out, log.StandardLogger())
		//fmt.Println(r)
		fmt.Printf("---\n\n")
	}

	return nil
}
