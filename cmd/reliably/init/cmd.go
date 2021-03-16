package init

import (
	"os"
	"strings"

	"github.com/reliablyhq/cli/core/manifest"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

var (
	manifestPath        string
	supportedExtensions = []string{".yaml", ".json"}
)

func NewCommand() *cobra.Command {
	cmd := cobra.Command{
		Use:   "init",
		Short: "initialise reliably",
		// Long: "Lorem ipsum...",
		Run: run,
	}

	cmd.Flags().StringVarP(&manifestPath, "manifest-file", "f", "reliably.yaml", "the location of the manifest file")

	return &cmd
}

func run(_ *cobra.Command, _ []string) {
	validateFilePath()

	m := manifest.Empty()

	// todo: ask some questions to populate the manifest

	f, err := os.Create(manifestPath)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	if err := yaml.NewEncoder(f).Encode(&m); err != nil {
		log.Fatal(err)
	}
}

func validateFilePath() {
	for _, ext := range supportedExtensions {
		if strings.HasSuffix(manifestPath, ext) {
			return
		}
	}

	log.Fatalf("manifest file must have one of the these extensions: %v", supportedExtensions)
}
