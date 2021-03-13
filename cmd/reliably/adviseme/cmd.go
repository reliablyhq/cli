package adviseme

import (
	"os"

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

	allAdvice := make([]*advice.Advice, 0)

	if a, err := advice.GetAdviceFor(m.Type); err != nil {
		log.Debug(err)
		log.Fatalf("an error occured while trying to get advice for '%s'", m.Type)
	} else {
		allAdvice = append(allAdvice, a)
	}

	if a, err := advice.GetAdviceFor(m.Platform); err != nil {
		log.Debug(err)
		log.Fatalf("an error occured while trying to get advice for '%s'", m.Type)
	} else {
		allAdvice = append(allAdvice, a)
	}

	if len(allAdvice) == 0 {
		log.Warn("no advice was found - how odd....")
		os.Exit(0)
	}

	for _, a := range allAdvice {
		for _, s := a.Suggestions {
			log.Infof("%s -> %s", a.Type, s)
		}
	}
}
