package auth

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/spf13/cobra"

	"github.com/AlecAivazis/survey/v2"

	"github.com/reliablyhq/cli/api"
	"github.com/reliablyhq/cli/config"
	"github.com/reliablyhq/cli/core/color"
	"github.com/reliablyhq/cli/core/iostreams"
)

// ResetLine returns the cursor to start of line and clean it
const ResetLine = "\r\033[K"

type LoginOptions struct {
	IO *iostreams.IOStreams

	Interactive bool

	Hostname string
	Token    string
}

// NewCmdLogin creates the `auth login` command
func NewCmdLogin() *cobra.Command {
	opts := &LoginOptions{
		IO: iostreams.System(),
	}

	var tokenStdin bool

	cmd := &cobra.Command{
		Use:   "login",
		Args:  cobra.ExactArgs(0),
		Short: "Authenticate with Reliably",
		Long: `Authenticate with Reliably.

The default authentication mode is interactive and asking for a token.

Alternatively, pass in a token on standard input by using '--with-token'.`,
		Example: heredoc.Doc(`
			# start interactive authentication setup
			$ reliably auth login

			# authenticate by reading the token from a file
			$ reliably auth login --with-token < my-access-token.txt`),

		PreRun: func(cmd *cobra.Command, args []string) {

		},

		RunE: func(cmd *cobra.Command, args []string) error {

			if tokenStdin {
				defer opts.IO.In.Close()
				token, err := io.ReadAll(opts.IO.In)
				if err != nil {
					return fmt.Errorf("failed to read token from STDIN: %w", err)
				}
				opts.Token = strings.TrimSpace(string(token))
			}

			if opts.IO.CanPrompt() && opts.Token == "" {
				opts.Interactive = true
			}

			if opts.Hostname == "" {
				opts.Hostname = config.Hostname
			}

			return loginRun(opts)
		},
	}

	cmd.Flags().StringVarP(&opts.Hostname, "hostname", "", "", "The hostname of Reliably to authenticate with")
	cmd.Flags().BoolVar(&tokenStdin, "with-token", false, "Read token from standard input")

	return cmd
}

func loginRun(opts *LoginOptions) error {
	hostname := opts.Hostname
	if hostname == "" {
		return errors.New("empty hostname")
	}

	if opts.Token != "" {
		return config.SetTokenForHostname(hostname, opts.Token)
	}

	if !opts.Interactive {
		return nil
	}

	fmt.Fprintf(opts.IO.ErrOut, "Logging into %s\n", hostname)

	// Check if a token already exists and is still valid
	if existingToken := config.GetTokenFor(hostname); existingToken != "" && opts.Interactive {
		apiClient := api.NewClientFromHTTP(api.AuthHTTPClient(hostname))
		if username, err := api.CurrentUsername(apiClient, hostname); err == nil {
			var keepGoing bool
			err = survey.AskOne(&survey.Confirm{
				Message: fmt.Sprintf(
					"You're already logged into %s as %s. Do you want to re-authenticate?",
					hostname,
					username),
				Default: false,
			}, &keepGoing, survey.WithShowCursor(true))
			if err != nil {
				return fmt.Errorf("could not prompt: %w", err)
			}

			if !keepGoing {
				return nil
			}
		}
	}

	var authMode int
	err := survey.AskOne(&survey.Select{
		Message: "How would you like to authenticate?",
		Options: []string{
			"Login with GitHub",
			"Login with GitLab",
			"Paste an authentication token",
		},
	}, &authMode)
	if err != nil {
		return fmt.Errorf("could not prompt for authentication mode: %w", err)
	}

	if authMode == 0 || authMode == 1 {

		var provider AuthProvider
		switch authMode {
		case 0:
			provider = AuthWithGithub
		case 1:
			provider = AuthWithGitlab
		}

		token, username, err := authFlow(hostname, provider)
		if err != nil {
			return fmt.Errorf("failed to authenticate via web browser: %w", err)
		}

		if err := config.SetAuthInfo(hostname, config.AuthInfo{Username: username, Token: token}); err != nil {
			return err
		}

		fmt.Fprintf(opts.IO.ErrOut, "%s Logged in as %s\n", iostreams.SuccessIcon(), color.Bold(username))

	} else {

		fmt.Fprintln(opts.IO.ErrOut)
		//fmt.Fprintln(opts.IO.ErrOut, heredoc.Doc(getAccessTokenTip(hostname)))

		var token string
		var username string
		err := survey.AskOne(&survey.Password{
			Message: "Paste your authentication token:",
		},
			&token,
			survey.WithShowCursor(true),
			survey.WithValidator(
				survey.ComposeValidators(
					survey.Required,
					func(val interface{}) error {
						// put token validation as part of the prompt,
						// user cannot pass the question with an invalid token

						token = val.(string)

						if err := config.SetTokenForHostname(hostname, token); err != nil {
							return err
						}

						// creates a new client that will use the token from config for hostname
						apiClient := api.NewClientFromHTTP(api.AuthHTTPClient(hostname))

						username, err = api.CurrentUsername(apiClient, hostname)
						if err != nil {
							if apiError, ok := err.(api.HTTPError); ok {
								if apiError.StatusCode == http.StatusUnauthorized {
									return fmt.Errorf("We were not able to identify you using the information you provided. Please make sure your token is valid.")
								}
							}
							return fmt.Errorf("error using api to retrieve user info: %w", err)
						}

						return nil
					},
				)))
		// forces start beginning on new line after prompt
		fmt.Fprint(opts.IO.ErrOut, ResetLine)
		if err != nil {
			return fmt.Errorf("could not prompt: %w", err)
		}

		if err := config.SetUsernameForHostname(hostname, username); err != nil {
			return err
		}

		fmt.Fprintf(opts.IO.ErrOut, "%s Logged in as %s\n", iostreams.SuccessIcon(), color.Bold(username))
	}

	return nil
}

/*
func getAccessTokenTip(hostname string) string {
	if hostname == "" {
		hostname = core.Hostname()
	}
	return fmt.Sprintf(`
	Tip: you can generate an Access Token here https://%s/tokens.`, hostname)
}
*/
