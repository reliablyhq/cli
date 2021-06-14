package sync

import (
	"errors"
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

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

func load(path string) ([]entities.Entity, error) {
	var objects []entities.Entity = make([]entities.Entity, 0)

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
	var m entities.Manifest
	if err := m.LoadFromFile(path); err != nil {
		return nil, err
	}

	for _, e := range m {
		objects = append(objects, e)
	}

	return objects, nil
}
