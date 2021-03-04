package cmd

import (
	"bytes"
	"text/template"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/spf13/cobra"

	"github.com/reliablyhq/cli/core/color"
)

var HelpTopics = map[string]map[string]string{
	"environment": {
		"short": "Environment variables that can be used with reliably",
		// Keeps generated lines at 80 characters max length
		"long": heredoc.Doc(`
			{{bold "RELIABLY_HOST:"}} specify the Reliably hostname for commands making API requests
			that would otherwise assume the "reliably.com" host.

			{{bold "RELIABLY_TOKEN:"}} an authentication token for reliably.com API requests. Setting
			this avoids to login and takes precedence over previously stored credentials.

			{{bold "RELIABLY_NO_UPDATE_NOTIFIER:"}} set to any value to disable update notifications.
			By default, reliably checks for new releases once every 24 hours and displays
			an upgrade notice on standard error if a newer version was found.
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

	// We use a template to be able to use coloring functions dynamically
	funcMap := template.FuncMap{
		"bold": color.Bold,
	}

	tmpl, err := template.New("").Funcs(funcMap).Parse(command.Long)
	if err != nil {
		er(err)
	}

	buf := new(bytes.Buffer)
	err = tmpl.Execute(buf, nil)
	if err != nil {
		er(err)
	}

	command.Print(buf.String())
}

func helpTopicUsageFunc(command *cobra.Command) error {
	command.Printf("Usage: reliably %s --help", command.Use)
	return nil
}
