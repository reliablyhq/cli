package report

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/reliablyhq/cli/core/color"
	"github.com/reliablyhq/cli/core/manifest"
	"github.com/reliablyhq/cli/core/report"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	manifestPath string
	outputPath   string
	outputFormat string
	watchFlag    bool
)

func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "report",
		Short: "Report my slo metrics",
		Run:   run,
	}

	cmd.Flags().StringVarP(&manifestPath, "manifest", "m", manifest.DefaultManifestPath, "the location of the manifest file")
	cmd.Flags().StringVarP(&outputPath, "output", "o", "", "where the report should be written to")
	cmd.Flags().StringVarP(&outputFormat, "format", "f", "tabbed", "specify the report format. Allowed Values: [json, simple, tabbed]")
	cmd.Flags().BoolVarP(&watchFlag, "watch", "w", false, "continously watch for changes in report output")

	return cmd
}

func run(_ *cobra.Command, _ []string) {

	// check for -w/--watch
	if watchFlag {
		watch(manifestPath)
		return
	}

	m, err := manifest.Load(manifestPath)
	if err != nil {
		log.Debug(err)

		if os.IsNotExist(err) {
			log.Fatal("A manifest was not found. Please run `reliably slo init` to create one.")
			return
		}

		log.Fatal("An error occured while attempting to load the manifest")
	}

	reports, err := report.FromManifest(m)
	if err != nil {
		log.Fatal(err)
	}

	// if err := sendReportToReliably(r); err != nil {
	// 	log.Warn(err)
	// }

	// set format
	var format = report.TABBED
	switch strings.ToLower(outputFormat) {
	case "json":
		format = report.JSON
	case "simple":
		format = report.SimpleText
	}

	for _, r := range reports {
		report.Write(format, r, os.Stdout, log.StandardLogger())
	}

	if outputPath != "" {
		if !strings.HasSuffix(outputPath, ".json") {
			log.Warn("output file should have a .json extension")
			return
		}

		bytes, err := json.Marshal(reports)
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

// watch - continously fetch and update report
// and output to terminal
func watch(manifestPath string) {
	rChan := make(chan []*report.Report, 5)
	c := make(chan os.Signal)
	done := make(chan struct{})
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	// refresh every 3 seconds
	go func() {
		for ch := time.Tick(time.Second * 3); ; <-ch {
			m, err := manifest.Load(manifestPath)
			if err != nil {
				log.Debug(err)

				if os.IsNotExist(err) {
					log.Fatal("A manifest was not found. Please run `reliably slo init` to create one.")
					return
				}

				log.Fatal("An error occured while attempting to load the manifest")
			}

			reports, err := report.FromManifest(m)
			if err != nil {
				log.Fatal(err)
			}

			rChan <- reports
		}
	}()

	// Ctrl+C listener
	go func() {
		<-c
		log.Info("CTRL+C pressed... exiting")
		// TODO: add any other cleanup actions here if needed
		done <- struct{}{}
	}()

	// print stuff
	for {

		select {
		case reports := <-rChan:
			clearScreen()
			fmt.Println(color.Magenta("Watching SLO report (3s)"))
			for _, r := range reports {
				report.Write(report.TABBED, r, os.Stdout, log.StandardLogger())
			}
			// tm.Flush()
		case <-done:
			return
		}
	}

}

func clearScreen() {
	var c *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		c = exec.Command("cmd", "/c", "cls")
	default:
		// clear should work for UNIX & linux based systems
		c = exec.Command("clear")
	}

	c.Stdout = os.Stdout
	c.Run()
}
