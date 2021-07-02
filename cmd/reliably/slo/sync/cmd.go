package sync

import (
	"bytes"
	"errors"
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/reliablyhq/cli/api"
	"github.com/reliablyhq/cli/config"
	"github.com/reliablyhq/cli/core/entities"
	"github.com/reliablyhq/cli/core/iostreams"
)

type SyncOptions struct {
	IO *iostreams.IOStreams

	ManifestPath string
}

var _ entities.Entity = &ManifestObject{}

type ManifestObject struct {
	APIVersion string                 `json:"apiVersion" yaml:"apiVersion"`
	KindValue  string                 `json:"kind" yaml:"kind"`
	Metadata   map[string]interface{} `json:"metadata" yaml:"metadata"`
	Spec       map[string]interface{} `json:"spec" yaml:"spec"`
}

func (t *ManifestObject) Version() string {
	return t.APIVersion
}

func (t *ManifestObject) Kind() string {
	return t.KindValue
}

func (t *ManifestObject) Validate() error {
	if t.APIVersion == "" {
		return errors.New("APIVersion is empty")
	}

	if t.KindValue == "" {
		return errors.New("kind is empty")
	}

	if len(t.Metadata) == 0 {
		return errors.New("metadata is invalid")
	}

	if len(t.Spec) == 0 {
		return errors.New("spec is invalid")
	}

	return nil
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
	entities, err := load(opts.ManifestPath)
	if err != nil {
		return err
	}

	hostname := config.Hostname
	entityHost := config.EntityServerHost
	apiClient := api.NewClientFromHTTP(api.AuthHTTPClient(hostname))

	org, err := config.GetCurrentOrgInfo()
	if err != nil {
		return err
	}

	var hasErr bool
	for i, entity := range entities {
		log.Debug("#", i, entity)
		err := api.CreateEntity(apiClient, entityHost, org.Name, entity)
		if err != nil {
			log.Debug(err)
			hasErr = true
		}
	}

	if hasErr {
		return errors.New("an error occured while syncing your manifest")
	} else {
		w := opts.IO.ErrOut
		fmt.Fprintln(w, iostreams.SuccessIcon(), "Your manifest has been successfully synchronized")
	}

	return nil
}

func load(path string) ([]entities.Entity, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	parts := bytes.Split(content, []byte("---"))
	entities := make([]entities.Entity, len(parts))
	for i, part := range parts {
		part = bytes.TrimSpace(part)

		var t ManifestObject
		if err := yaml.Unmarshal(part, &t); err != nil {
			return nil, err
		}

		if err := t.Validate(); err != nil {
			return nil, err
		}

		entities[i] = &t
	}

	return entities, nil
}
