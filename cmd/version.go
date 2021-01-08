package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of the Reliably CLI",
	Long:  `Print the version number of the Reliably CLI`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Print(FormatVersion(version))
	},
	Hidden: true, // version will not be indicated as command but only flag
}

// FormatVersion returns a formatted string with Reliably CLI's version
func FormatVersion(version string) string {
	version = strings.TrimPrefix(version, "v")
	return fmt.Sprintf("Reliably CLI v%s\n", version)
}
