package auth

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"
	//"github.com/spf13/viper"

	"github.com/reliablyhq/cli/api"
	"github.com/reliablyhq/cli/core"
	"github.com/reliablyhq/cli/core/color"
	"github.com/reliablyhq/cli/core/config"
	"github.com/reliablyhq/cli/core/iostreams"
)

type StatusOptions struct {
	IO *iostreams.IOStreams

	Hostname  string
	ShowToken bool
}

var silentError = errors.New("")

// NewCmdStatus creates the `auth status` command
func NewCmdStatus() *cobra.Command {
	opts := &StatusOptions{
		IO: iostreams.System(),
	}

	cmd := &cobra.Command{
		Use:   "status",
		Args:  cobra.ExactArgs(0),
		Short: "View authentication status",
		Long: `Verifies and displays information about your authentication to Reliably.

This command will test your authentication token to ensure
it's still valid and report on any issue.`,

		PreRun: func(cmd *cobra.Command, args []string) {

		},

		RunE: func(cmd *cobra.Command, args []string) error {

			if opts.Hostname == "" {
				opts.Hostname = core.Hostname()
			}

			return statusRun(opts)
		},
	}

	cmd.Flags().StringVarP(&opts.Hostname, "hostname", "", "", "Check a specific hostname's auth status")
	cmd.Flags().BoolVarP(&opts.ShowToken, "show-token", "t", false, "Display the auth token")

	return cmd
}

func statusRun(opts *StatusOptions) error {
	/*
		statusInfo := map[string][]string{}
		statusInfo[hostname] = []string{}
		addMsg := func(x string, ys ...interface{}) {
			statusInfo[hostname] = append(statusInfo[hostname], fmt.Sprintf(x, ys...))
		}
	*/

	stderr := opts.IO.ErrOut
	hostname := opts.Hostname

	statusInfo := []string{}
	addMsg := func(x string, ys ...interface{}) {
		statusInfo = append(statusInfo, fmt.Sprintf(x, ys...))
	}

	var err error
	var username string
	var token string
	var loginCmd string = "reliably auth login"
	var logoutCmd string = "reliably auth logout"

	var loginHostCmd string = loginCmd
	var logoutHostCmd string = logoutCmd
	if hostname != core.Hostname() {
		loginHostCmd = fmt.Sprintf("%s --hostname %s", loginCmd, hostname)
		logoutHostCmd = fmt.Sprintf("%s --hostname %s", logoutCmd, hostname)
	}

	if len(config.Hosts()) == 0 {
		return fmt.Errorf(
			"You are not logged into Reliably. Run '%s' to authenticate", loginCmd)
	}

	// need to ensure authentication exists in config for hostname
	tokenIsWriteable := core.AuthTokenProvidedFromEnv() == false
	token, env, err := config.GetAuthTokenWithSource(hostname)
	if err != nil {
		return fmt.Errorf(
			"You are not logged in to %s. Run '%s' to authenticate",
			hostname, loginHostCmd)
	}

	var failed bool

	apiClient := api.NewClientFromHTTP(api.AuthHTTPClient(hostname))
	username, err = api.CurrentUsername(apiClient, hostname)
	if err != nil {
		if err, ok := err.(api.HTTPError); ok && err.StatusCode == 401 {
			addMsg(fmt.Sprintf("%s %s", iostreams.FailureIcon(), "authentication failed"))
			addMsg("- The access token in %s is no longer valid.", hostname)
			if tokenIsWriteable {
				addMsg("- To re-authenticate, run: %s", loginHostCmd)
				addMsg("- To forget about this authentication, run: %s", logoutHostCmd)
			}
			failed = true
		} else {
			return fmt.Errorf("Unable to check token validity against %s", hostname)
		}

	} else {

		usernameStr := ""
		if username != "" {
			usernameStr = fmt.Sprintf(" as %s", string(username))
		}
		addMsg("%s Logged in to %s%s (%s)", iostreams.SuccessIcon(), hostname, usernameStr, env)

		tokenDisplay := "*******************"
		if opts.ShowToken {
			tokenDisplay = token
		}
		addMsg("%s Token: %s", iostreams.SuccessIcon(), tokenDisplay)
	}

	fmt.Fprintf(stderr, "%s\n", color.Bold(hostname))
	for _, line := range statusInfo {
		fmt.Fprintf(stderr, "  %s\n", line)
	}

	if failed {
		return silentError
	}

	return nil

}
