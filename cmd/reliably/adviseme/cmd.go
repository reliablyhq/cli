package adviseme

import (
	"os"

	"github.com/reliablyhq/cli/core"
	"github.com/reliablyhq/cli/core/advice"
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

	cmd.Flags().StringVarP(&manifestPath, "manifest-path", "p", "", "the path to the manifest file")

	return cmd
}

func run(_ *cobra.Command, _ []string) {
	m, err := core.LoadManifest(manifestPath)
	if err != nil {
		log.Debug(err)

		if err == os.ErrExist {
			log.Fatal("A manifest was not found. Please run `reliably init` to create one.")
			return
		}

		log.Fatal("An error occured while attempting to load the manifest")
	}

	allAdvice, err := advice.GetAdviceFor(m)
	if err != nil {
		log.Debug(err)
		log.Fatal("an error occured while geting advice")
	}

	for _, a := range allAdvice {
		for _, s := range a.Suggestions {
			log.Infof("%s -> %s", a.Type, s)
		}
	}
}
