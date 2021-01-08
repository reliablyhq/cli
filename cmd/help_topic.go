package cmd

import (
	"fmt"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/spf13/cobra"
)

var HelpTopics = map[string]map[string]string{
	"environment": {
		"short": "Environment variables that can be used with gh",
		"long": heredoc.Doc(`
			RELIABLY_HOST: specify the Reliably hostname for commands making
			API requests that would otherwise assume the "reliably.com" host.

			RELIABLY_TOKEN: an authentication token for reliably.com API
			requests. Setting this avoids to login and takes precedence over
			previously stored credentials.
		`),
	},
}

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

func helpTopicHelpFunc(command *cobra.Command, args []string) {
	command.Print(command.Long)
}

func helpTopicUsageFunc(command *cobra.Command) error {
	fmt.Println("test")
	command.Printf("Usage: reliably %s --help", command.Use)
	return nil
}
