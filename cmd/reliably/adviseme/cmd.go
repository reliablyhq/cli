package adviseme

import (
	"github.com/sirupsen/logrus"
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
	logrus.Fatal("not implemented")
}
