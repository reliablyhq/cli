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

	"github.com/reliablyhq/cli/api"
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
	org          string
	service      string
)

func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "report",
		Short: "Report my slo metrics",
		RunE:  runE,
	}

	cmd.Flags().StringVarP(&manifestPath, "manifest", "m", manifest.DefaultManifestPath, "the location of the manifest file")
	cmd.Flags().StringVarP(&outputPath, "output", "o", "", "where the report should be written to")
	cmd.Flags().StringVarP(&outputFormat, "format", "f", "tabbed", "specify the report format. Allowed Values: [json, simple, tabbed, markdown]")
	cmd.Flags().BoolVarP(&watchFlag, "watch", "w", false, "continuously watch for changes in report output")
	cmd.Flags().StringVar(&org, "org", "", "the org that contains the service")

	return cmd
}

func runE(_ *cobra.Command, _ []string) error {

	// check for -w/--watch
	if watchFlag {
		return watch(manifestPath)
	}

	m, err := getManifest()
	if err != nil {
		log.Debug(err)

		if os.IsNotExist(err) {
			return errors.New("A manifest was not found. Please run `reliably slo init` to create one.")
		}

		return errors.New("An error occured while attempting to load the manifest")
	}

	r, err := report.FromManifest(m)
	if err != nil {
		return err
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
	case "markdown":
		format = report.MARKDOWN
	}

	report.Write(format, r, os.Stdout, log.StandardLogger())

	if outputPath != "" {
		if !strings.HasSuffix(outputPath, ".json") {
			log.Warn("output file should have a .json extension")
			return nil
		}

		bytes, err := json.Marshal(r)
		if err != nil {
			return nil
		}

		if err := ioutil.WriteFile(outputPath, bytes, 0666); err != nil {
			return err
		}
	}

	return nil
}

func sendReportToReliably(r *report.Report) error {
	return errors.New("Sending reports to Reliably is not available yet. Check back later :D")
}

// watch - continously fetch and update report
// and output to terminal
func watch(manifestPath string) error {
	rChan := make(chan *report.Report, 5)
	errChan := make(chan error, 1)
	done := make(chan struct{})
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	defer func() {
		// put cursor back on return
		fmt.Print("\033[?25h")
	}()

	// refresh every 3 seconds
	go func() {
		for ch := time.Tick(time.Second * 3); ; <-ch {
			m, err := manifest.Load(manifestPath)
			if err != nil {
				log.Debug(err)
				if os.IsNotExist(err) {
					errChan <- errors.New("A manifest was not found. Please run `reliably slo init` to create one.")
					return
				}
				errChan <- errors.New("An error occured while attempting to load the manifest")
			}

			report, err := report.FromManifest(m)
			if err != nil {
				errChan <- err
			}
			rChan <- report
		}
	}()

	// Ctrl+C listener
	go func() {
		<-c
		fmt.Printf("\nCTRL+C pressed... exiting\n")
		done <- struct{}{}
	}()

	// print stuff
	for {
		select {
		case r := <-rChan:
			clearScreen()
			fmt.Println(color.Magenta("Refreshing SLO report every 3 seconds."), "Press CTRL+C to quit.")
			report.Write(report.TABBED, r, os.Stdout, log.StandardLogger())

		case err := <-errChan:
			return err

		case <-done:
			return nil
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

		// hide cursor on unix based systems
		fmt.Print("\033[?25l")
	}

	c.Stdout = os.Stdout
	c.Run()
}

func getManifest() (*manifest.Manifest, error) {
	if org != "" && service != "" {
		return api.PullManifest(org, service)
	}

	return manifest.Load(manifestPath)
}
