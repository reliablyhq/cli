package auth

import (
	"fmt"

	"github.com/spf13/cobra"
	//"github.com/spf13/viper"
	"github.com/AlecAivazis/survey/v2"

	"github.com/reliablyhq/cli/core"
	"github.com/reliablyhq/cli/core/config"
	"github.com/reliablyhq/cli/core/iostreams"
)

type LogoutOptions struct {
	IO *iostreams.IOStreams

	Interactive bool

	Hostname  string
	NoConfirm bool
}

// NewCmdLogout creates the `auth logout` command
func NewCmdLogout() *cobra.Command {
	opts := &LogoutOptions{
		IO: iostreams.System(),
	}

	cmd := &cobra.Command{
		Use:   "logout",
		Args:  cobra.ExactArgs(0),
		Short: "Log out from Reliably",
		Long: `Remove authentication for Reliably.

This command removes the authentication configuration.`,

		PreRun: func(cmd *cobra.Command, args []string) {

		},

		RunE: func(cmd *cobra.Command, args []string) error {

			if opts.Hostname == "" {
				opts.Hostname = core.Hostname()
			}

			return logoutRun(opts)
		},
	}

	cmd.Flags().StringVarP(&opts.Hostname, "hostname", "", "", "The hostname of Reliably to log out of")
	cmd.Flags().BoolVarP(&opts.NoConfirm, "yes", "y", false, "Don't ask for logout confirmation")

	return cmd
}

func logoutRun(opts *LogoutOptions) error {
	hostname := opts.Hostname
	askConfirm := !opts.NoConfirm

	auths := config.Viper.GetStringMap("auths")

	// need to ensure authentication exists in config for hostname
	if auths[hostname] == nil {
		return fmt.Errorf("You are not logged in to %s", hostname)
	}

	username := config.Viper.Get(fmt.Sprintf("auths::%s::username", hostname))

	usernameStr := ""
	if username != "" {
		usernameStr = fmt.Sprintf(" account '%s'", username)
	}

	if opts.IO.CanPrompt() && askConfirm {
		var keepGoing bool
		err := survey.AskOne(&survey.Confirm{
			Message: fmt.Sprintf("Are you sure you want to log out of %s%s?", hostname, usernameStr),
			Default: true,
		}, &keepGoing)
		if err != nil {
			return fmt.Errorf("could not prompt: %w", err)
		}

		if !keepGoing {
			return nil
		}
	}

	// do stuff to remove host from config
	delete(auths, hostname)
	config.Viper.Set("auths", auths)

	// save updated config
	err := config.Viper.WriteConfig()
	if err != nil {
		return fmt.Errorf("failed to write config, authentication configuration not updated: %w", err)
	}

	isTTY := opts.IO.IsStdinTTY() && opts.IO.IsStdoutTTY()
	if isTTY {
		fmt.Printf("Logged out of %s%s\n", hostname, usernameStr)
	}

	return nil
}
