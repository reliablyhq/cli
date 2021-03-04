package terraform

import (
	"github.com/reliablyhq/cli/cmd/reliably/scan/terraform/plan"
	"github.com/spf13/cobra"
)

// New terraform command
func New() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "terraform",
		Short: "scan terraform resources",
		Long:  "scan terraform resources for policy violations",
	}

	cmd.AddCommand(plan.New())

	return cmd
}
