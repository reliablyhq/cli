package list

import (
	"io"

	"github.com/olekukonko/tablewriter"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/reliablyhq/cli/api"
	"github.com/reliablyhq/cli/config"
	"github.com/reliablyhq/cli/core/color"
	"github.com/reliablyhq/cli/core/iostreams"
)

type ListOptions struct {
	IO *iostreams.IOStreams
}

func Command() *cobra.Command {

	opts := &ListOptions{
		IO: iostreams.System(),
	}

	return &cobra.Command{
		Use:   "list",
		Short: "list organizations ",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runList(opts)
		},
	}
}

func runList(opts *ListOptions) error {
	hostname := config.Hostname

	activeOrg, err := config.GetCurrentOrgInfo()
	if err != nil {
		return err
	}

	client := api.NewClientFromHTTP(api.AuthHTTPClient(hostname))

	userID, err := api.CurrentUserID(client, hostname)
	if err != nil {
		return err
	}

	orgs, err := api.ListOrganizations(client, hostname)
	if err != nil {
		return err
	}

	log.Debugf("List of organizations with member '%s': %s", "<USERNAME>", orgs)

	return renderOrgsList(opts.IO.Out, orgs, userID, activeOrg.ID)
}

func renderOrgsList(w io.Writer, orgs []api.Organization, ownerID string, currentOrgID string) error {

	table := tablewriter.NewWriter(w)
	table.SetAutoFormatHeaders(false)
	//table.SetAutoWrapText(true)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetBorder(false)
	table.SetRowLine(false)
	table.SetRowSeparator("")
	table.SetColumnSeparator("")
	table.SetHeaderLine(false)
	//table.SetColWidth(maxColWidth)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetHeader([]string{
		"",     //current active org from config
		"Name", // organization name
		"ID",   // organization ID
		"",     // user relationship to organization: member or owner
	})

	for _, org := range orgs {
		active := " "
		if currentOrgID == org.ID {
			active = "*"
		}
		userRelation := color.Grey("member")
		if ownerID == org.Owner || (org.Owner == "" && ownerID == org.CreatedBy) {
			userRelation = color.Yellow("owner")

		}

		row := []string{active, color.Bold(org.Name), org.ID, userRelation}
		table.Append(row)

	}
	table.Render()

	return nil
}
