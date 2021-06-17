package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/reliablyhq/cli/cmd/reliably/cmdutil"
	v "github.com/reliablyhq/cli/version"
)

func NewCmdVersion() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print the version number of the Reliably CLI",
		Long:  `Print the version number of the Reliably CLI`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Print(FormatVersion(version, buildDate, revision))
		},
		Hidden: true, // version will not be indicated as command but only flag
	}

	cmdutil.DisableAuthCheck(cmd)

	return cmd
}

// FormatVersion returns a formatted string with Reliably CLI's version
func FormatVersion(version string, buildDate string, revision string) string {
	version = strings.TrimPrefix(version, "v")
	var buildDateStr string = ""
	var gitRevStr string = ""

	if buildDate != "" {
		buildDateStr = fmt.Sprintf(" (%s)", buildDate)
	}

	if version == v.DevVersion && revision != "" {
		gitRevStr = fmt.Sprintf(" (rev %s)", revision)
	}

	return fmt.Sprintf("Reliably CLI version %s%s%s\n", version, buildDateStr, gitRevStr)
}
