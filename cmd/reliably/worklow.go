package cmd

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/MakeNowJust/heredoc/v2"
	"github.com/spf13/cobra"

	"github.com/reliablyhq/cli/core/iostreams"
	wf "github.com/reliablyhq/cli/core/workflows"
	"github.com/reliablyhq/cli/utils"
)

type WorkflowOptions struct {
	IO *iostreams.IOStreams

	Interactive bool
	Stdout      bool

	Platform string
}

var (
	supportedCIPlatforms = Choice{"github", "gitlab", "circleci"}
)

func NewCmdWorkflow() *cobra.Command {
	return newCmdWorkflow(nil)
}

func newCmdWorkflow(runF func(opts *WorkflowOptions) error) *cobra.Command {

	opts := &WorkflowOptions{
		IO: iostreams.System(),
	}

	cmd := &cobra.Command{
		Use:   "workflow",
		Short: "Setup your Reliably workflow",
		Long: `Setup the Reliably workflow for your CI/CD platform.

It runs in an interactive mode by default.`,

		Example: heredoc.Doc(`
			# Setup your workflow
			$ reliably workflow

			# Run in non-interactive mode, by specifying the platform as argument
			$ reliably workflow --platform=github
		`),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if opts.Platform != "" && !supportedCIPlatforms.Has(opts.Platform) {
				return fmt.Errorf("Platform '%v' is not valid. Use one of the supported platforms: %v", opts.Platform, supportedCIPlatforms)
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {

			if opts.IO.CanPrompt() && opts.Platform == "" {
				opts.Interactive = true
			}

			if runF == nil {
				return workflowRun(opts)
			}

			return runF(opts)
		},
	}

	cmd.Flags().StringVarP(
		&opts.Platform, "platform", "p", "",
		fmt.Sprintf("Specify the CI/CD platform: %v", supportedCIPlatforms))
	cmd.Flags().BoolVar(
		&opts.Stdout, "stdout", false,
		"Generate the workflow to standard output rather than file")
	return cmd
}

func workflowRun(opts *WorkflowOptions) error {

	if opts.Interactive {
		var selectedPlatform string
		err := survey.AskOne(&survey.Select{
			Message: "Select your CI/CD platform to create the Reliably workflow",
			Options: []string{
				"GitHub",
				"GitLab",
				"CircleCI",
			},
		}, &selectedPlatform)
		if err != nil {
			return fmt.Errorf("could not prompt for CI/CD platform: %w", err)
		}

		// Makes sure the choice is converted to
		// for now, simpler as it is, lowercase value of the human friendly name
		opts.Platform = strings.ToLower(selectedPlatform)
	}

	platform := opts.Platform

	if opts.Stdout {
		// show the workflow to standard output, rather than writing to file
		fmt.Fprintf(opts.IO.Out, "%s\n", wf.GetWorkflow(opts.Platform))
		return nil
	}

	wfPath := wf.GetWorkflowPath(platform)
	if _, err := os.Stat(wfPath); !os.IsNotExist(err) && !wf.CanEditWorkflowInPlace(platform) {
		return fmt.Errorf("A Reliably workflow already exists at path: %s", wfPath)
	}

	wfPath, err := wf.GenerateWorkflow(platform)
	if err != nil {
		return err
	}
	fmt.Fprintf(opts.IO.ErrOut, "Your workflow has been generated to %s\n", wfPath)

	if utils.IsGitRepo() {
		worfklowGitHelp(opts.IO.ErrOut, wfPath)
	}

	workflowAccessTokenHelp(opts.IO.ErrOut, platform)

	return nil
}

// worfklowGitHelp prints out the git commands to add & commit
// the workflow file to the current repository
func worfklowGitHelp(w io.Writer, path string) {

	fmt.Fprintf(w, `
You can now add and commit the workflow to your repository:
$ git add %s
$ git commit -m "Add Reliably workflow"
`,
		path)

}

// workflowAccessTokenHelp prints out the help message on how to
// setup securely the Reliably access token as RELIABLY_TOKEN env var
// for any CI/CD platform.
func workflowAccessTokenHelp(w io.Writer, platform string) {
	help := wf.GetAccessTokenHelp(platform, "RELIABLY_TOKEN")
	if help != "" {
		fmt.Fprint(w, "\n", help, "\n",
			"You can retrieve your access token by running:\n",
			"$ reliably auth status --show-token\n")
	}
}
