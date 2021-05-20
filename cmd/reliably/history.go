package cmd

import (
	//"errors"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/AlecAivazis/survey/v2/terminal"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/reliablyhq/cli/api"
	"github.com/reliablyhq/cli/core"
	"github.com/reliablyhq/cli/core/color"
	ctx "github.com/reliablyhq/cli/core/context"
	"github.com/reliablyhq/cli/core/iostreams"
	"github.com/reliablyhq/cli/core/output"
	"github.com/reliablyhq/cli/utils"
)

const ResetLine = "\r\033[K"
const ClearPreviousLine = "\033[1A\033[K"

type HistoryOptions struct {
	IO         *iostreams.IOStreams
	Hostname   string
	HttpClient func() *http.Client

	History  *interface{}
	OrgID    string
	SourceID string
}

func NewCmdHistory() *cobra.Command {
	opts := &HistoryOptions{
		IO:       iostreams.System(),
		Hostname: core.Hostname(),
		HttpClient: func() *http.Client {
			return api.AuthHTTPClient(core.Hostname())
		},
	}

	cmd := &cobra.Command{
		Use:   "history",
		Short: "Show your scan history",
		Long:  `Show your entire history of executions and found suggestions.`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			var err error
			apiClient := api.NewClientFromHTTP(opts.HttpClient())

			// Ensure the CLI history is executed in a valid org/source
			opts.OrgID, err = api.CurrentUserOrganizationID(apiClient, opts.Hostname)
			if err != nil {
				return fmt.Errorf("unable to retrieve current organization: %w", err)
			}

			context := ctx.NewContext() // can we improve/refactor to create a source without full context
			opts.SourceID, err = api.CurrentSourceID(apiClient, opts.Hostname, opts.OrgID, context.Source.(ctx.Source).Hash)
			if err != nil {
				if e, ok := err.(api.HTTPError); ok {
					if e.StatusCode == 404 {
						return fmt.Errorf(`Current source is unknown!
You probably haven't run 'reliably scan .' from this current working directory yet.
The history will only be available after a first scan.`)
					}
				}
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

	log.WithFields(log.Fields{
		"org":    opts.OrgID,
		"source": opts.SourceID,
	}).Debug("Run 'history' command with")

	apiClient := api.NewClientFromHTTP(opts.HttpClient())

	var (
		cursor       string
		hasNext      bool = true
		noHistory    bool = true
		currentExec  string
		currentCount int // suggestions of an execution might be split on two+ pages
	)

	for hasNext {
		history, err := api.GetSuggestionHistory(apiClient, opts.Hostname, opts.OrgID, opts.SourceID, cursor)
		if err != nil {
			return fmt.Errorf("Unable to retrieve your execution & suggestion history: %w", err)
		}

		if history == nil {
			break
		}

		if len(history.Executions) > 0 {
			noHistory = false
		}

		for _, exec := range history.Executions {

			// we start a new execution - print out the header
			if exec.ID != currentExec {
				if currentExec != "" {
					// prints the footer for previous exec - except for first one
					printExecFooter(opts, nil, currentCount)
				}

				printExecHeader(opts, &exec)
				currentExec = exec.ID
				currentCount = len(exec.Suggestions)
			} else {
				// we're continuing the exection on a different cursor/page
				currentCount = currentCount + len(exec.Suggestions)
			}

			printExecSuggestions(opts, &exec)

			// at end of the current execution, we don't know whether the execution
			// has more suggestions to come in the next cursor-based pagination
		}

		hasNext = history.PageInfo.HasNextPage
		cursor = history.PageInfo.Cursor

		//hasNext = false // For DEV only @TODO remove

		if !hasNext {
			// prints out latest exec footer, before leaving
			printExecFooter(opts, nil, currentCount)
			fmt.Fprintln(opts.IO.ErrOut, "You reached the end of your history!")
		}

		if hasNext {
			// prompt user before loading more
			err := promptToLoadMore(opts)
			if err != nil {
				if err == terminal.InterruptErr {
					os.Exit(0)
				}
				return err
			}
		}

	}

	// handle final message when no history was found for current source
	if noHistory {
		return errors.New("You have no history yet!")
	}

	return nil

}

func printExecHeader(opts *HistoryOptions, exec *api.Execution) {
	header := color.Yellow(fmt.Sprintf("Execution %s", exec.ID))
	sub := fmt.Sprintf("Date: %s", exec.Date.Format(time.RFC1123))

	fmt.Fprintln(opts.IO.Out, header)
	fmt.Fprintln(opts.IO.Out, sub)
	fmt.Fprintln(opts.IO.Out)
}

func printExecSuggestions(opts *HistoryOptions, exec *api.Execution) {

	var suggestions []*core.Suggestion
	for _, s := range exec.Suggestions {
		suggestions = append(suggestions, s.Data)
	}

	output.CreateReport(opts.IO.Out, "simple", "", suggestions)
}

func printExecFooter(opts *HistoryOptions, exec *api.Execution, nbSuggestions int) {
	if nbSuggestions > 0 {
		plural := utils.IfThenElse(nbSuggestions > 1, "s", "")
		msg := color.Red(fmt.Sprintf("%v suggestion%s found", nbSuggestions, plural))
		fmt.Fprintf(opts.IO.Out, "%s %s\n", iostreams.FailureIcon(), msg)
	} else {
		msg := color.Green("No suggestion found!")
		fmt.Fprintf(opts.IO.Out, "%s %s\n", iostreams.SuccessIcon(), msg)
	}
	fmt.Fprintln(opts.IO.Out)
}

func promptToLoadMore(opts *HistoryOptions) error {

	null := ""

	prompt := &survey.Input{
		Message: "press ENTER to load more entries or Ctrl+C to exit...",
		//Help:    "You have more entries in your history.",
	}

	err := survey.AskOne(prompt, &null, survey.WithIcons(func(icons *survey.IconSet) {
		icons.Question.Text = ">"
		icons.Question.Format = "green"
	}), survey.WithStdio(os.Stdin, os.Stderr, os.Stderr), survey.WithShowCursor(false)) // we redirect prompt to stderr only !
	// i wanted to be able to use opts.IO.x but not same type for survey IO :(

	if err == nil {
		// clear prompted message, to continue printing next to latest history line
		// (as enter has been pressed ie cursor is on a new line)
		fmt.Fprint(opts.IO.ErrOut, ClearPreviousLine)
	} else {
		if err == terminal.InterruptErr {
			// when ctrl-c, cursor is still on the prompt line
			fmt.Fprint(opts.IO.ErrOut, ResetLine)
		}
	}

	return err
}
