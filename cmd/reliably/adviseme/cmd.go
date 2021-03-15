package adviseme

import (
	"os"

	"github.com/reliablyhq/cli/core/advice"
	"github.com/reliablyhq/cli/manifest"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	manifestPath string
)

func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "adviseme",
		Short: "Get some advice",
		Run:   run,
	}

	cmd.Flags().StringVarP(&manifestPath, "manifest-file", "f", "", "the path to the manifest file")

	return cmd
}

func run(_ *cobra.Command, _ []string) {
	m, err := manifest.Load(manifestPath)
	if err != nil {
		log.Debug(err)

		if err == os.ErrExist {
			log.Fatal("A manifest was not found. Please run `reliably init` to create one.")
			return
		}

		log.Fatal("An error occured while attempting to load the manifest")
	}

	log.Debug("manifest: ", m)

	allAdvice, err := advice.GetAdviceFor(m)
	if err != nil {
		log.Error(err)
	}

	for _, s := range allAdvice.Suggestions {
		log.Info(s)
	}
}
