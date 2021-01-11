package cmd

import (
	"github.com/spf13/cobra"

	authCmd "github.com/reliablyhq/cli/cmd/auth"
)

func NewCmdAuth() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth <command>",
		Short: "Login, logout, and verify your authentication",
		Long:  `Manage Reliably's authentication state.`,
	}

	cmd.AddCommand(authCmd.NewCmdLogin())
	cmd.AddCommand(authCmd.NewCmdLogout())
	cmd.AddCommand(authCmd.NewCmdStatus())

	return cmd
}

func init() {
	rootCmd.AddCommand(NewCmdAuth())
}
