package cmd

import (
	"github.com/spf13/cobra"
)

func NewCmdDiscover() *cobra.Command {
	cmd := &cobra.Command{
		Use:        "discover [path]",
		Hidden:     true,
		Deprecated: "Please use `reliably scan` instead",
	}

	return cmd
}
