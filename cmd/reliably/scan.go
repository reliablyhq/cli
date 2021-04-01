package cmd

import (
	"github.com/spf13/cobra"

	"github.com/reliablyhq/cli/cmd/reliably/scan"
)

func NewCmdScan(rootCmd *cobra.Command) *cobra.Command {

	cmd := &cobra.Command{
		Use:   "scan [resource]",
		Short: "Check for Reliably Suggestions",
		//Long: `Check your manifests for Reliably Suggestions.`,
	}

	cmd.AddCommand(scan.NewCmdScanK8s(rootCmd))

	return cmd
}
