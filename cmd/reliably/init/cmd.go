package init

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func NewCommand() *cobra.Command {
	cmd := cobra.Command{
		Use:   "init",
		Short: "initialise reliably",
		// Long: "Lorem ipsum...",
		Run: run,
	}

	return &cmd
}

func run(_ *cobra.Command, _ []string) {
	log.Fatal("not implemented :(")
}
