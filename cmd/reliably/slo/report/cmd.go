package report

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/reliablyhq/cli/api"
	"github.com/reliablyhq/cli/cmd/reliably/cmdutil"
	"github.com/reliablyhq/cli/core"
	"github.com/reliablyhq/cli/core/color"
	"github.com/reliablyhq/cli/core/iostreams"
	"github.com/reliablyhq/cli/core/manifest"
	"github.com/reliablyhq/cli/core/report"
)

type Choice = cmdutil.Choice

type ReportOptions struct {
	IO *iostreams.IOStreams
}

var (
	supportedFormats  = Choice{"json", "yaml", "text", "table", "markdown"}
	deprecatedFormats = Choice{"simple", "tabbed"}
	manifestPath      string
	outputPath        string
	outputFormat      string
	watchFlag         bool

	service string
)

func NewCommand() *cobra.Command {
	opts := &ReportOptions{
		IO: iostreams.System(),
	}

	cmd := &cobra.Command{
		Use:   "report",
		Short: "Report my slo metrics",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			if !cmdutil.CheckAuth() {
				cmdutil.PrintRequireAuthMsg()
				os.Exit(1)
			}
		},
		PreRunE: func(cmd *cobra.Command, args []string) error {
			// Validate command options
			if outputFormat != "" && !(supportedFormats.Has(outputFormat) || deprecatedFormats.Has(outputFormat)) {
				return fmt.Errorf("Format '%v' is not valid. Use one of the supported formats: %v", outputFormat, supportedFormats)
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return reportRun(opts)
		},
	}

	cmd.Flags().StringVarP(&manifestPath, "manifest", "m", manifest.DefaultManifestPath, "the location of the manifest file")
	cmd.Flags().StringVarP(&outputPath, "output", "o", "", "where the report should be written to")
	cmd.Flags().StringVarP(&outputFormat, "format", "f", "table", fmt.Sprintf("specify the report format. Allowed Values: %v", supportedFormats))
	cmd.Flags().BoolVarP(&watchFlag, "watch", "w", false, "continuously watch for changes in report output")
	cmd.Flags().StringVar(&service, "service", "", "the name of the service")

	return cmd
}

func reportRun(opts *ReportOptions) error {
	// check for -w/--watch
	if watchFlag {
		return watch()
	}

	if outputFormat != "" && deprecatedFormats.Has(outputFormat) {
		log.Warnf("Format '%v' is now deprecated and soon be to removed. Use one of the supported formats: %v", outputFormat, supportedFormats)
	}

	opts.IO.StartProgressIndicator()

	m, err := getManifest()
	if err != nil {
		log.Debug(err)
		return errors.New("an error occured while attempting to load the manifest")
	}

	if m == nil {
		return errors.New("no service manifest detected")
	}

	r, err := report.FromManifest(m)
	if err != nil {
		return err
	}

	apiClient := api.NewClientFromHTTP(api.AuthHTTPClient(core.Hostname()))
	orgID, _ := api.CurrentUserOrganizationID(apiClient, core.Hostname())
	if _, err := api.SendReport(apiClient, orgID, r); err != nil {
		log.Debugf("Error while sending report to reliably: %s", err)
	}

	// set format
	var format = report.TABBED
	switch strings.ToLower(outputFormat) {
	case "json":
		format = report.JSON
	case "simple", "text":
		format = report.SimpleText
	case "markdown":
		format = report.MARKDOWN
	case "yaml":
		format = report.YAML
	}

	var w io.Writer = os.Stdout
	if outputPath != "" {
		outfile, err := os.Create(outputPath) // creates or truncates with O_RDWR mode
		if err != nil {
			log.Error("error creating output file")
			log.Error(err)
			return err
		}
		w = outfile
		defer outfile.Close()
	}

	opts.IO.StopProgressIndicator()

	report.Write(format, r, w, log.StandardLogger())

	return nil
}

// watch - continously fetch and update report
// and output to terminal
func watch() error {
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
			m, err := getManifest()
			if err != nil {
				log.Debug(err)
				errChan <- errors.New("an error occured while attempting to load the manifest")
			}

			if m == nil {
				errChan <- errors.New("no service manifest detected")
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

// getManifest priority
// 1. local file
// 2. service manifest - if specified
// 3 full manifest download
func getManifest() (m *manifest.Manifest, err error) {

	m, err = manifest.Load(manifestPath)
	if err == nil {
		if service != "" {
			var services []*manifest.Service
			for _, s := range m.Services {
				if s.Name == service {
					services = append(services, s)
				}
			}

			m.Services = services
		}
		return
	}

	log.Debugf("unable to read manifest file: %s - %s", manifestPath, err)
	log.Debug("attempting to retrieve manifest from reliably api")

	client := api.NewClientFromHTTP(api.AuthHTTPClient(core.Hostname()))
	if service == "" {
		m, err = api.PullManifest(client)
		return
	}

	return api.PullServiceManifest(client, service)
}
