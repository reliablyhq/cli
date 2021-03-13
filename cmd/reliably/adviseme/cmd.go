package adviseme

import (
	"github.com/reliablyhq/cli/core"
	"github.com/reliablyhq/cli/core/advice"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func NewCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "adviseme",
		Short: "Get some advice",
		Run:   run,
	}
}

func run(_ *cobra.Command, _ []string) {
	m, err := core.LoadManifest()
	if err != nil {
		log.Debug(err)
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
