package pull

import (
	"os"

	"github.com/reliablyhq/cli/api"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

var output string

func NewCommand() *cobra.Command {
	cmd := cobra.Command{
		Use:   "pull",
		Short: "pull/download SLO manifest from reliably",
		// Long:    longCommandDescription(),
		// Example: examples(),
		RunE: runE,
	}

	cmd.Flags().StringVarP(&output, "output", "o", "reliably.yaml", "output path/location manifest file will be stored")
	return &cmd
}

func runE(cmd *cobra.Command, args []string) error {
	log.Debugf("pulling manifest to: [%s]", output)
	m, err := api.PullServiceManifest("", "")
	if err != nil {
		return err
	}

	f, err := os.Create(output)
	if err != nil {
		return err
	}
	defer f.Close()

	if err := yaml.NewEncoder(f).Encode(&m); err != nil {
		return err
	}
	return nil
}
