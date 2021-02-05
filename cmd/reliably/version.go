package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(NewCmdVersion())
}

func NewCmdVersion() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print the version number of the Reliably CLI",
		Long:  `Print the version number of the Reliably CLI`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Print(FormatVersion(version, buildDate))
		},
		Hidden: true, // version will not be indicated as command but only flag
	}

	return cmd
}

// FormatVersion returns a formatted string with Reliably CLI's version
func FormatVersion(version string, buildDate string) string {
	version = strings.TrimPrefix(version, "v")
	var buildDateStr string = ""

	if buildDate != "" {
		buildDateStr = fmt.Sprintf(" (%s)", buildDate)
	}

	return fmt.Sprintf("Reliably CLI version %s%s\n", version, buildDateStr)
}
