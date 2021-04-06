package cmdutil

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/reliablyhq/cli/core/color"
)

// CustomUsageTemplate returns usage template for the command.
// This is the default usage template from the command,
// with the additional help:feedback annotations, if specified
func CustomUsageTemplate(c *cobra.Command) string {

	tpl := c.UsageTemplate()

	if _, ok := c.Annotations["help:environment"]; ok {
		environment := c.Annotations["help:environment"]
		tpl = fmt.Sprintf("%s%s", tpl, environment)
	}

	if _, ok := c.Annotations["help:feedback"]; ok {
		feedback := c.Annotations["help:feedback"]
		tpl = fmt.Sprintf("%s%s", tpl, feedback)
	}

	cobra.AddTemplateFunc("StyleHeading", color.Yellow)
	replacer := strings.NewReplacer(
		`Usage:`, `{{StyleHeading "Usage:"}}`,
		`Aliases:`, `{{StyleHeading "Aliases:"}}`,
		`Examples:`, `{{StyleHeading "Examples:"}}`,
		`Available Commands:`, `{{StyleHeading "Available Commands:"}}`,
		`Global Flags:`, `{{StyleHeading "Global Flags:"}}`,
		// The following one steps on "Global Flags:"
		`Flags:`, `{{StyleHeading "Flags:"}}`,
		`Environment variables:`, `{{StyleHeading "Environment variables:"}}`,
		`Feedback:`, `{{StyleHeading "Feedback:"}}`,
		`Additional help topics:`, `{{StyleHeading "Additional help topics:"}}`,
	)
	tpl = replacer.Replace(tpl)

	return tpl
}
