package cmd

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/reliablyhq/cli/core/iostreams"
)

func NewCmdCompletion() *cobra.Command {

	var shellType string
	var io *iostreams.IOStreams = iostreams.System()

	cmd := &cobra.Command{
		Use:   "completion -s <shell>",
		Short: "Generate shell completion scripts",
		Long: `Generate shell completion scripts for Reliably CLI commands.

The output of this command will be computer code and is meant to be saved to a
file or immediately evaluated by an interactive shell.

For example, for bash you could add this to your '~/.bash_profile':
  eval "$(reliably completion -s bash)"`,
		Hidden:                false,
		DisableFlagsInUseLine: true,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if shellType == "" {
				if io.IsStdoutTTY() {
					return errors.New("the value for `--shell` is required")
					//return &cmdutil.FlagError{Err: errors.New("error: the value for `--shell` is required")}
				}
				shellType = "bash"
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {

			w := io.Out
			rootCmd := cmd.Parent()

			switch shellType {
			case "bash":
				return rootCmd.GenBashCompletion(w)
			case "zsh":
				return rootCmd.GenZshCompletion(w)
			case "powershell":
				return rootCmd.GenPowerShellCompletion(w)
			case "fish":
				return rootCmd.GenFishCompletion(w, true)
			default:
				return fmt.Errorf("unsupported shell type %q", shellType)
			}
		},
	}

	cmd.Flags().StringVarP(&shellType, "shell", "s", "", "Shell type: {bash|zsh|fish|powershell}")

	return cmd
}

/*
func NewHelpTopic(topic string) *cobra.Command {

	cmd := &cobra.Command{
		Use:    topic,
		Short:  HelpTopics[topic]["short"],
		Long:   HelpTopics[topic]["long"],
		Hidden: true,
		Annotations: map[string]string{
			"markdown:generate": "true",
			"markdown:basename": "help_" + topic,
		},
	}

	cmd.SetHelpFunc(helpTopicHelpFunc)
	cmd.SetUsageFunc(helpTopicUsageFunc)

	return cmd
}
*/
