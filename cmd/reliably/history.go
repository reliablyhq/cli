package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/reliablyhq/cli/api"
	"github.com/reliablyhq/cli/cmd/reliably/cmdutil"
	"github.com/reliablyhq/cli/core"
	ctx "github.com/reliablyhq/cli/core/context"
	"github.com/reliablyhq/cli/core/iostreams"
)

type HistoryOptions struct {
	IO        *iostreams.IOStreams
	Hostname  string
	ApiClient *api.Client

	History  *interface{}
	SourceID string
}

func NewCmdHistory() *cobra.Command {
	opts := &HistoryOptions{
		IO:       iostreams.System(),
		Hostname: core.Hostname(),
	}

	cmd := &cobra.Command{
		Use:   "history",
		Short: "Show your discovery history",
		Long:  `Show your entire history of executions and found suggestions.`,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			if !cmdutil.CheckAuth() {
				cmdutil.PrintRequireAuthMsg()
				os.Exit(1)
			}
		},
		PreRunE: func(cmd *cobra.Command, args []string) error {
			// cannot be done when creating the opts map, as
			// the init config has not been called yet !
			// instead we need to ensure the command is initialized  properly
			// so that we can read from the config
			opts.ApiClient = api.NewClientFromHTTP(api.AuthHTTPClient(opts.Hostname))

			// Ensure the CLI history is executed in a valid org/source
			orgID, err := api.CurrentUserOrganizationID(opts.ApiClient, opts.Hostname)
			if err != nil {
				return fmt.Errorf("unable to retrieve current organization: %w", err)
			}

			context := ctx.NewContext() // can we improve/refactor to create a source without full context
			opts.SourceID, err = api.CurrentSourceID(opts.ApiClient, opts.Hostname, orgID, context.Source.(ctx.Source).Hash)
			if err != nil {
				return fmt.Errorf("unable to retrieve current Source: %w", err)
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return historyRun(opts)
		},
	}

	return cmd
}

func historyRun(opts *HistoryOptions) (err error) {

	fmt.Printf("Showing history for Source %s\n", opts.SourceID)

	return nil

}
