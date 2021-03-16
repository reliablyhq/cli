package report

import (
	"errors"
	"os"

	"github.com/reliablyhq/cli/core/manifest"
	"github.com/reliablyhq/cli/core/report"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	manifestPath string
)

func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "report",
		Short: "Report my resiliency metrics",
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

	r, err := report.GenerateReport(m)
	if err != nil {
		log.Error(err)
	}

	sendReportToReliably(r)

	report.Write(r, log.StandardLogger())
}

func sendReportToReliably(r *report.Report) error {
	return errors.New("not implemented")
}
