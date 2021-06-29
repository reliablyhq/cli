package create

import (
	"errors"
	"fmt"
	"net/http"
	"regexp"

	"github.com/AlecAivazis/survey/v2"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/reliablyhq/cli/api"
	"github.com/reliablyhq/cli/config"
	"github.com/reliablyhq/cli/core/cli/question"
	"github.com/reliablyhq/cli/core/iostreams"
)

const orgNameConvention = "^[a-zA-Z0-9]+(?:-[a-zA-Z0-9]+)*$"

type CreateOptions struct {
	IO *iostreams.IOStreams

	Interactive bool
	Name        string
}

func Command() *cobra.Command {

	opts := &CreateOptions{
		IO: iostreams.System(),
	}

	cmd := &cobra.Command{
		Use:   "create [<name>]",
		Short: "create a new organization",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				opts.Name = args[0]
			} else {
				opts.Interactive = true
			}

			return runCreate(opts)
		},
	}

	//	cmd.Flags().StringVarP(&orgName, "name", "n", "", "the name of the organization to create")

	return cmd
}

func runCreate(opts *CreateOptions) error {

	hostname := config.Hostname

	tryCreate := true
	for tryCreate { // this loop allows to prompt again, in interactive mode, when org names are already used

		if opts.Interactive {
			opts.Name = interactiveOrgName()
		}

		client := api.NewClientFromHTTP(api.AuthHTTPClient(hostname))
		err := api.CreateOrganisation(client, hostname, opts.Name)
		if err != nil {
			log.Debug(err)

			if err, ok := err.(api.HTTPError); ok {
				if err.StatusCode == http.StatusConflict {
					msg := fmt.Sprintf("Orgnization name '%s' is not available.", opts.Name)
					if !opts.Interactive {
						return errors.New(msg)
					} else {
						fmt.Fprintln(opts.IO.Out, iostreams.FailureIcon(), msg, "Please choose another.")
					}
				}
			}

			if !opts.Interactive {
				// un-expected error, returns directly
				return errors.New("An unexpected error ocurred while creating the organization")
			}

		} else {
			tryCreate = false // exit loop once org is created
		}

	}

	msg := fmt.Sprintf("organization '%s' created", opts.Name)
	fmt.Fprintln(opts.IO.Out, iostreams.SuccessIcon(), msg)

	return nil
}

func interactiveOrgName() string {

	name := question.WithStringAnswerV2(
		"What is the name of the organization?",
		"Name may only contain alphanumeric characters or single hyphens, and cannot begin or end with a hyphen.",
		"",
		[]question.AskOpt{
			survey.WithValidator(
				survey.ComposeValidators(
					survey.Required,
					func(val interface{}) error {
						// Ensure the org name follows the convention
						name := val.(string)
						re, _ := regexp.Compile(orgNameConvention)
						if !re.MatchString(name) {
							return errors.New("organization name is not valid, may only contain alphanumeric characters or single hyphens, and cannot begin or end with a hyphen.")
						}
						return nil
					},
					func(val interface{}) error {
						// enture the org name has maximum length
						name := val.(string)
						if len(name) >= 40 {
							return errors.New("organization name is too long (maximum is 39 characters).")
						}
						return nil
					},
				)),
		})

	return name
}
