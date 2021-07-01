package sync

import (
	"errors"
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/reliablyhq/cli/api"
	"github.com/reliablyhq/cli/config"
	"github.com/reliablyhq/cli/core/entities"
	"github.com/reliablyhq/cli/core/iostreams"
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
		Short: "synchronize the objectives from your manifest with Reliably",
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
	var m entities.Manifest
	if err := m.LoadFromFile(opts.ManifestPath); err != nil {
		return err
	}

	hostname := config.Hostname
	entityHost := config.EntityServerHost
	apiClient := api.NewClientFromHTTP(api.AuthHTTPClient(hostname))

	org, err := config.GetCurrentOrgInfo()
	if err != nil {
		return err
	}

	if err := api.SyncManifest(apiClient, entityHost, org.Name, m); err != nil {
		return err
	}

	w := opts.IO.ErrOut
	fmt.Fprintln(w, iostreams.SuccessIcon(), "Your manifest has been successfully synchronized")
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
