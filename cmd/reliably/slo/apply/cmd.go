package apply

import (
	"errors"
	"fmt"

	"github.com/reliablyhq/cli/api"
	"github.com/reliablyhq/cli/core/manifest"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var manifestPath string

func NewCommand() *cobra.Command {
	cmd := cobra.Command{
		Use:   "apply",
		Short: "Apply a service manifest to organization by filename",
		// Long:    longCommandDescription(),
		// Example: examples(),
		RunE: runE,
	}

	cmd.Flags().StringVarP(&manifestPath, "file", "f", "", "store a local copy of the service manifest created")
	return &cmd
}

func runE(cmd *cobra.Command, args []string) error {
	if manifestPath == "" {
		return errors.New("please specify manifest file, using -f/--file")
	}

	log.Debugf("pushing manifest: [%s]", manifestPath)
	m, err := manifest.Load(manifestPath)
	if err != nil {
		return err
	}

	if err := api.PushManifest(m); err != nil {
		return fmt.Errorf("an error occurred while push manifest to reliably: %s", err)
	}
	return nil
}
