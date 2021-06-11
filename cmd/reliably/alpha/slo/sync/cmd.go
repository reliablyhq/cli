package sync

import (
	"errors"
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"

	"github.com/reliablyhq/cli/api"
	"github.com/reliablyhq/cli/core"
	"github.com/reliablyhq/cli/core/color"
	"github.com/reliablyhq/cli/core/entities"
	"github.com/reliablyhq/cli/core/iostreams"
	v "github.com/reliablyhq/cli/version"
)

type SyncOptions struct {
	IO *iostreams.IOStreams

	ManifestPath string
}

func NewCommand(runF func(*SyncOptions) error) *cobra.Command {
	opts := &SyncOptions{
		IO: iostreams.System(),
	}

	cmd := cobra.Command{
		Use:   "sync",
		Short: "synchronize your manifest with our servers",
		RunE: func(cmd *cobra.Command, args []string) error {
			if runF != nil {
				return runF(opts)
			}

			return syncRun(opts)
		},
	}

	cmd.Flags().StringVarP(&opts.ManifestPath, "manifest", "m", "reliably.yaml", "path of the manifest to sync up")
	return &cmd
}

func syncRun(opts *SyncOptions) error {
	objectives, err := load(opts.ManifestPath)
	if err != nil {
		return err
	}

	hostname := core.Hostname()
	entityHost := core.Hostname()
	if v.IsDevVersion() {
		if hostFromEnv := os.Getenv("RELIABLY_ENTITY_HOST"); hostFromEnv != "" {
			entityHost = hostFromEnv
		}
	}

	apiClient := api.NewClientFromHTTP(api.AuthHTTPClient(hostname))
	org, _ := api.CurrentUserOrganization(apiClient, hostname)

	var hasErr bool
	for i, slo := range objectives {
		log.Debug("#", i, slo)
		err := api.CreateEntity(apiClient, entityHost, org.Name, slo)
		if err != nil {
			log.Debug(err)
			hasErr = true
		}
	}

	if hasErr {
		return errors.New("An error occured while syncing your manifest")
	} else {
		w := opts.IO.ErrOut
		fmt.Fprintln(w, iostreams.SuccessIcon(), color.Green("Your manifest has been successfully synchronized"))
	}

	return nil
}

func load(path string) ([]entities.Objective, error) {
	var objects []entities.Objective = make([]entities.Objective, 0)

	if path == "" {
		return nil, errors.New("path is empty")
	}

	log.Debug("Loading manifest at ", path)

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	//var m Manifest
	dec := yaml.NewDecoder(file)

	var objective *entities.Objective
	for dec.Decode(&objective) == nil {
		objects = append(objects, *objective)
		// ensure to create a new pointer for next iteration - avoid merged sub-props
		objective = new(entities.Objective)
	}

	return objects, nil
}
