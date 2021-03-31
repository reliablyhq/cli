package report

import (
	"encoding/json"
	"errors"
	"io/ioutil"
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
		Short: "Report my slo metrics",
		Run:   run,
	}

	cmd.Flags().StringVarP(&manifestPath, "manifest-file", "f", manifest.DefaultManifestPath, "the path to the manifest file")
	cmd.Flags().StringVarP(&outputPath, "output", "o", "", "where the report should be written to")

	return cmd
}

func run(_ *cobra.Command, _ []string) {
	m, err := manifest.Load(manifestPath)
	if err != nil {
		log.Debug(err)

		if os.IsNotExist(err) {
			log.Fatal("A manifest was not found. Please run `reliably slo init` to create one.")
			return
		}

		log.Fatal("An error occured while attempting to load the manifest")
	}

	r, err := report.FromManifest(m)
	if err != nil {
		log.Fatal(err)
	}

	if err := sendReportToReliably(r); err != nil {
		log.Warn(err)
	}

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

		if err := ioutil.WriteFile(outputPath, bytes, 0666); err != nil {
			log.Fatal(err)
		}
	}
}

func sendReportToReliably(r *report.Report) error {
	return errors.New("Sending reports to Reliably is not available yet. Check back later :D")
}
