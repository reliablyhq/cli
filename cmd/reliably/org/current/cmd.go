package current

import (
	"fmt"

	//log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/reliablyhq/cli/config"
	"github.com/reliablyhq/cli/core/color"
	"github.com/reliablyhq/cli/core/iostreams"
)

type ShowOptions struct {
	IO *iostreams.IOStreams
}

func Command() *cobra.Command {

	opts := &ShowOptions{
		IO: iostreams.System(),
	}

	return &cobra.Command{
		Use:   "current",
		Short: "show the organization defined as current",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runShowCurrent(opts)
		},
	}
}

func runShowCurrent(opts *ShowOptions) error {

	info, err := config.GetCurrentOrgInfo()
	if err != nil {
		return err
	}

	orgName := info.Name
	orgID := info.ID

	if orgName == "" {
		orgName = "UNKNOWN"
	}

	if orgID == "" {
		orgID = "UNKNOWN"
	}

	w := opts.IO.Out
	indent := "\t"
	fmt.Fprintln(w, "Current organization:")
	fmt.Fprintln(w, indent, "Name:", indent, color.Bold(orgName))
	fmt.Fprintln(w, indent, "ID:", indent, orgID)

	return nil
}
