package plan

import (
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	file string
)

// New returns a new plan command
func New() *cobra.Command {
	cmd := cobra.Command{
		Use:   "plan",
		Short: "scan a terraform plan",
		Long:  "scan a terraform plan for policy violations",
		Run:   run,
	}

	cmd.Flags().StringVarP(&file, "file", "f", "", "the path to the terraform plan")

	return &cmd
}

func run(cmd *cobra.Command, args []string) {
	if file == "" {
		log.Error("file argument is required")
		os.Exit(1)
	}

	// 1: check for the file

	// 2: parse the file content

	// 3: find policies for each resource

	// 4: analyse

	// 4: print the outcome
	log.Warn("Not implemented!")
	os.Exit(1)
}
