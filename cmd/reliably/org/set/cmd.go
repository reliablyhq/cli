package set

import (
	"errors"
	"fmt"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/reliablyhq/cli/api"
	"github.com/reliablyhq/cli/config"
	"github.com/reliablyhq/cli/core/cli/question"
	"github.com/reliablyhq/cli/core/iostreams"
)

type SetOptions struct {
	IO *iostreams.IOStreams

	Interactive bool
	Name        string
}

func Command() *cobra.Command {

	opts := &SetOptions{
		IO: iostreams.System(),
	}

	cmd := &cobra.Command{
		Use:   "set [<name>]",
		Short: "defines an organization as the current one",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				opts.Name = args[0]
			} else {
				opts.Interactive = true
			}

			return runSetOrg(opts)
		},
	}

	return cmd
}

func runSetOrg(opts *SetOptions) error {

	// 1: get current orgs for user
	hostname := config.Hostname

	log.Debug("getting orgs for current user from ", hostname)
	client := api.NewClientFromHTTP(api.AuthHTTPClient(hostname))
	orgs, err := api.ListOrganizations(client, hostname)
	if err != nil {
		return err
	}
	log.Debugf("got %d orgs", len(orgs))

	// 2: prompt for interactive org selection
	if opts.Interactive {
		var orgNames []string
		for _, o := range orgs {
			orgNames = append(orgNames, o.Name)
		}
		opts.Name = question.WithSingleChoiceAnswer(
			"Select the organization to set as current",
			[]survey.AskOpt{},
			orgNames...)
	}

	// 3: check if org is in that collection - for non interactive -
	var org *api.Organization
	for _, o := range orgs {
		if strings.EqualFold(o.Name, opts.Name) {
			org = &o
			break
		}
	}

	if org == nil {
		return errors.New("The organization does not exist or you are not a member of it")
	}
	log.Debug("matching org found - updating config file...")

	// 3: write org to config
	fmt.Println(org.Name, org.ID)
	if err := config.SetCurrentOrgInfo(org.Name, org.ID); err != nil {
		return err
	}

	msg := fmt.Sprintf("Config file has been updated with organization '%s' as default", opts.Name)
	fmt.Fprintln(opts.IO.Out, iostreams.SuccessIcon(), msg)

	return nil
}
