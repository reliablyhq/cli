package report

import (
	"encoding/json"
	"errors"
	"os"
	"strings"

	"github.com/reliablyhq/cli/core/manifest"
	"github.com/reliablyhq/cli/core/report"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	manifestPath string
	outputPath   string
)

func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "report",
		Short: "Report my resiliency metrics",
		Run:   run,
	}

	cmd.Flags().StringVarP(&manifestPath, "manifest-file", "f", "", "the path to the manifest file")
	cmd.Flags().StringVarP(&outputPath, "output", "o", "", "where the report should be written to")

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

	if outputPath != "" {
		if !strings.HasSuffix(outputPath, ".json") {
			log.Warn("output file should have a .json extension")
			return
		}

		bytes, err := json.Marshal(r)
		if err != nil {
			log.Fatal(err)
		}

		if err := os.WriteFile(outputPath, bytes, 0666); err != nil {
			log.Fatal(err)
		}
	}
}

func sendReportToReliably(r *report.Report) error {
	return errors.New("not implemented")
}
