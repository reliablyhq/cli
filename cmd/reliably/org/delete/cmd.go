package delete

import (
	"errors"
	"fmt"

	"github.com/AlecAivazis/survey/v2"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/reliablyhq/cli/api"
	"github.com/reliablyhq/cli/cmd/reliably/org/shared"
	"github.com/reliablyhq/cli/config"
	"github.com/reliablyhq/cli/core/cli/question"
	"github.com/reliablyhq/cli/core/iostreams"
)

type DeleteOptions struct {
	IO *iostreams.IOStreams

	Interactive bool
	Name        string
}

func Command() *cobra.Command {

	opts := &DeleteOptions{
		IO: iostreams.System(),
	}

	cmd := &cobra.Command{
		Use:   "delete [<name>]",
		Short: "delete an organization",
		Long: `Delete an existing organization.

You can only delete an organization for which you are the owner.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				opts.Name = args[0]
			} else {
				opts.Interactive = true
			}

			return runDelete(opts)
		},
	}

	return cmd
}

func runDelete(opts *DeleteOptions) error {

	hostname := config.Hostname
	log.Debug("getting orgs for current user from ", hostname)

	client := api.NewClientFromHTTP(api.AuthHTTPClient(hostname))

	userID, err := api.CurrentUserID(client, hostname)
	if err != nil {
		return err
	}

	orgs, err := api.ListOrganizations(client, hostname)
	if err != nil {
		return err
	}

	owned := shared.FilterOrgs(orgs, func(o api.Organization) bool {
		return shared.IsOwner(userID, o)
	})

	if opts.Interactive {
		var orgNames []string
		for _, o := range owned {
			orgNames = append(orgNames, o.Name)
		}
		opts.Name = question.WithSingleChoiceAnswer(
			"Select the organization to delete",
			[]survey.AskOpt{},
			orgNames...)
	} else {
		// ensure org name exists
		org := shared.FilterOrgByName(&orgs, opts.Name)
		if org == nil {
			return fmt.Errorf("organization '%s' not found", opts.Name)
		}
	}

	// ensure current user is owner ie can delete
	org := shared.FilterOrgByName(&owned, opts.Name)
	if org == nil {
		return fmt.Errorf("organization '%s' cannot be deleted (only owner can delete)", opts.Name)
	}

	if err := api.DeleteOrganisation(client, hostname, org.ID); err != nil {
		log.Debug(err)
		return errors.New("An unexpected error ocurred while deleting the organization")
	}

	msg := fmt.Sprintf("organization '%s' deleted", opts.Name)
	fmt.Fprintln(opts.IO.Out, iostreams.SuccessIcon(), msg)

	return nil
}

/*
// filterOrgs returns the list of organizations that succeeded the test function
func filterOrgs(orgs []api.Organization, test func(o api.Organization) bool) (ret []api.Organization) {
	for _, o := range orgs {
		if test(o) {
			ret = append(ret, o)
		}
	}
	return
}

func isOwner(ID string, org api.Organization) bool {
	return ID == org.Owner || (org.Owner == "" && ID == org.CreatedBy)
}

func filterOrgByName(orgs *[]api.Organization, name string) *api.Organization {
	var org *api.Organization
	for _, o := range *orgs {
		if strings.EqualFold(name, o.Name) {
			org = &o
			break
		}
	}
	return org
}
*/
