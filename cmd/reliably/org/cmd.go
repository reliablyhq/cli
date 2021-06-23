package org

import (
	"github.com/MakeNowJust/heredoc/v2"
	"github.com/spf13/cobra"

	"github.com/reliablyhq/cli/cmd/reliably/org/add_user"
	"github.com/reliablyhq/cli/cmd/reliably/org/create"
	"github.com/reliablyhq/cli/cmd/reliably/org/current"
	"github.com/reliablyhq/cli/cmd/reliably/org/delete"
	"github.com/reliablyhq/cli/cmd/reliably/org/list"
	"github.com/reliablyhq/cli/cmd/reliably/org/remove_user"
	"github.com/reliablyhq/cli/cmd/reliably/org/set"
)

func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "org",
		Short: "Manage your organizations",
		Long: heredoc.Doc(`
			A set of commands to manage your organizations and their members.

			Organizations allows you to regroup and share data between multiple users.
			You can create, manage and delete organizations.

			You can also add/remove users to/from your organizations.`),
	}

	cmd.AddCommand(
		create.Command(),
		current.Command(),
		list.Command(),
		set.Command(),
		add_user.Command(),
		remove_user.Command(),
		delete.Command(),
	)

	return cmd
}
