package add_user

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/MakeNowJust/heredoc/v2"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/reliablyhq/cli/api"
	"github.com/reliablyhq/cli/cmd/reliably/org/shared"
	"github.com/reliablyhq/cli/config"
	"github.com/reliablyhq/cli/core/iostreams"
)

type AddOptions struct {
	IO *iostreams.IOStreams

	Username string
	OrgName  string
}

func Command() *cobra.Command {

	opts := &AddOptions{
		IO: iostreams.System(),
	}

	cmd := &cobra.Command{
		Use:   "add-user <username>",
		Short: "add a user to the current organization",
		Long: heredoc.Docf(`
			Add a user as a member of an existing organization.

			By default, the user is added to the current organization.
			It is possible to add a user to a specific organization,
			provided with the %[1]s--org%[1]s flag.

			Only the owner of the organization is able to manages users.`, "`"),
		Example: heredoc.Doc(`
			# add a user to the current organization
			$ reliably org add-user <someone-user-name>

			# add a user to a specific organization
			$ reliably org add-user <someone-user-name> --org <my-other-org>`),
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				opts.Username = args[0]
			} else {
				return errors.New("required username argument is missing")
			}

			return runAddUserToOrg(opts)
		},
	}

	cmd.Flags().StringVarP(&opts.OrgName, "org", "o", "", "specify the organization to which add the user")

	return cmd
}

func runAddUserToOrg(opts *AddOptions) error {

	hostname := config.Hostname
	client := api.NewClientFromHTTP(api.AuthHTTPClient(hostname))

	username := opts.Username

	current, err := config.GetCurrentOrgInfo()
	if err != nil {
		return err
	}

	if current.ID == "" && opts.OrgName == "" {
		return errors.New("The current organization has not been set. Please run `reliably org set`.")
	}

	orgID := current.ID
	orgName := current.Name

	if opts.OrgName != "" {
		// if specified, use org from cmd flags in precedence over default org
		orgs, err := api.ListOrganizations(client, hostname)
		if err != nil {
			return err
		}

		org := shared.FilterOrgByName(&orgs, opts.OrgName)
		if org == nil {
			return fmt.Errorf("Organization '%s' not found", opts.OrgName)
		}

		orgID = org.ID
		orgName = org.Name
	}

	if err := api.AddUserToOrganisation(client, hostname, orgID, username); err != nil {
		// check API status codes
		// 404 -> "org not found"
		// 404 -> "user not found"
		// 403 -> not member of the org -> no permission
		// 409 -> user is already a member
		log.Debug(err)
		if err, ok := err.(api.HTTPError); ok {
			switch err.StatusCode {
			case http.StatusNotFound:
				if strings.Contains(err.Message, "org not found") {
					return fmt.Errorf("Organization '%s' not found", orgName)
				}
				if strings.Contains(err.Message, "user not found") {
					return fmt.Errorf("No user found with username '%s'", username)
				}
			case http.StatusForbidden:
				return errors.New("You are not allowed to add users into this organization. Ensure you are either the owner or a member.")
			case http.StatusConflict:
				return fmt.Errorf("User '%s' is already a member of the organization '%s'", username, orgName)
			}

		}
		// unknown error - return it
		return err
	}

	msg := fmt.Sprintf("user '%s' added to organization '%s'", username, orgName)
	fmt.Fprintln(opts.IO.Out, iostreams.SuccessIcon(), msg)

	return nil
}
