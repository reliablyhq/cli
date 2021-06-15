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

	"github.com/MakeNowJust/heredoc/v2"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/reliablyhq/cli/api"
	"github.com/reliablyhq/cli/cmd/reliably/cmdutil"
	"github.com/reliablyhq/cli/config"
	"github.com/reliablyhq/cli/core/color"
	"github.com/reliablyhq/cli/core/iostreams"
	"github.com/reliablyhq/cli/core/manifest"
	"github.com/reliablyhq/cli/core/report"
	"github.com/reliablyhq/cli/utils"
)

type Choice = cmdutil.Choice

type ReportOutput struct {
	Format report.Format
	Path   string
}

type ReportOptions struct {
	IO *iostreams.IOStreams

	ManifestPath  string
	OutputPath    string
	OutputFormat  string
	TemplateFile  string
	WatchFlag     bool
	OutputPaths   []string
	OutputFormats []string
	Service       string

	Outputs []ReportOutput
}

const defaultFormat = "table"

var (
	supportedFormats  = Choice{"json", "yaml", "text", "table", "markdown", "template"}
	deprecatedFormats = Choice{"simple", "tabbed"}
)

func NewCommand(runF func(*ReportOptions) error) *cobra.Command {
	opts := &ReportOptions{
		IO: iostreams.System(),
	}

	cmd := &cobra.Command{
		Use:   "report",
		Short: "Report my slo metrics",
		Long: heredoc.Doc(`Generates a report of your SLOs.

It is also possible to generate the report to different files &
formats at once, with using '--format' and '--output' flags with
comma-separated list as values.`),
		Example: `  $ reliably slo report
  $ reliably slo report -f text
  $ reliably slo report -f markdown -o report.md
  $ reliably slo report -f yaml,json -o o.yaml,o.json
  $ reliably slo report -t slo-report.tmpl
  $ reliably slo report -t slo-report.tmpl -o slo-report.txt`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			// Validate command options

			if opts.TemplateFile != "" {
				// check template file exists, see outputFormat to template
				_, err := os.Open(opts.TemplateFile) // For read access.
				if err != nil {
					return fmt.Errorf("Error opening template file: %s", err)
				}

				opts.OutputFormat = "template"
			}

			if opts.OutputFormat != "" {
				opts.OutputFormats = strings.Split(opts.OutputFormat, ",")
			}

			if len(opts.OutputFormats) > 0 {
				for _, of := range opts.OutputFormats {
					if of != "" && !(supportedFormats.Has(of) || deprecatedFormats.Has(of)) {
						return fmt.Errorf("Format '%v' is not valid. Use one of the supported formats: %v", of, supportedFormats)
					}
				}
			}

			for _, of := range opts.OutputFormats {
				if of != "" && deprecatedFormats.Has(of) {
					log.Warnf("Format '%v' is now deprecated and soon be to removed. Use one of the supported formats: %v", of, supportedFormats)
				}
			}

			if opts.OutputPath != "" {
				opts.OutputPaths = strings.Split(opts.OutputPath, ",")
			}

			if len(opts.OutputFormats) > 1 && len(opts.OutputPaths) == 0 {
				return errors.New("Multiple output formats must be used in combination with multiple output path '--output o1,o2,...' flag")
			}

			if len(opts.OutputFormats) == 1 && opts.OutputFormat == defaultFormat && len(opts.OutputPaths) > 1 {
				return errors.New("Each output file specified with '--output' must have a format defined with '--format f1,f2,...'")
			}

			if len(opts.OutputFormats) > 0 && len(opts.OutputPaths) > 0 &&
				len(opts.OutputFormats) != len(opts.OutputPaths) {
				return errors.New("Flags '--format' and '--output' must have same number of values when combined")
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {

			// given the list of formats & outputs,
			// create the list of ReportOutput structs to combine a format and (optional) path
			// --format f1,f2 --output o1,o2 should be associated as (f1, o1) & (f2, o2)
			for fIdx, of := range opts.OutputFormats {
				var format = report.TABBED
				switch strings.ToLower(of) {
				case "json":
					format = report.JSON
				case "simple", "text":
					format = report.SimpleText
				case "markdown":
					format = report.MARKDOWN
				case "template":
					format = report.TEMPLATE
				case "yaml":
					format = report.YAML
				}

				// get the output path at same index
				var path string
				if len(opts.OutputPaths) > fIdx {
					path = opts.OutputPaths[fIdx]
				}

				opts.Outputs = append(opts.Outputs, ReportOutput{
					Format: format,
					Path:   path,
				})

			}

			if runF != nil {
				return runF(opts)
			}

			return reportRun(opts)
		},
	}

	cmd.Flags().StringVarP(&opts.ManifestPath, "manifest", "m", manifest.DefaultManifestPath, "the location of the manifest file")
	cmd.Flags().StringVarP(&opts.OutputPath, "output", "o", "", "where the report should be written to")
	cmd.Flags().StringVarP(&opts.TemplateFile, "template", "t", "", "the name of the template to use for the report output")
	cmd.Flags().StringVarP(&opts.OutputFormat, "format", "f", "table", fmt.Sprintf("specify the report format. Allowed Values: %v", supportedFormats))
	cmd.Flags().BoolVarP(&opts.WatchFlag, "watch", "w", false, "continuously watch for changes in report output")
	cmd.Flags().StringVar(&opts.Service, "service", "", "the name of the service")

	return cmd
}

func reportRun(opts *ReportOptions) error {
	// check for -w/--watch
	if opts.WatchFlag {
		return watch(opts)
	}

	opts.IO.StartProgressIndicator()

	hostname := config.Hostname
	apiClient := api.NewClientFromHTTP(api.AuthHTTPClient(hostname))
	orgID, _ := api.CurrentUserOrganizationID(apiClient, hostname)

	// TODO refactoring, this is very sequential right now, we could
	// improve by using a bit of paralellism with goroutines

	// ! we need to fetch the last report before pushing the new one !
	var lr report.Report
	var reports []report.Report
	var err error
	if reports, err = api.GetReports(apiClient, hostname, orgID, 4); err == nil {
		if len(reports) > 0 {
			lr = reports[0]
		}
	}

	m, err := getManifest(opts.ManifestPath, opts.Service)
	if err != nil {
		log.Debug(err)
		return errors.New("an error occurred while attempting to load the manifest")
	}

	if m == nil {
		return errors.New("no service manifest detected")
	}

	r, err := report.FromManifest(m)
	if err != nil {
		return err
	}

	// reverse last reports - oldest to most recent - append current at the end
	utils.Reverse(reports)
	reports = append(reports, *r)

	//apiClient := api.NewClientFromHTTP(api.AuthHTTPClient(core.Hostname()))
	//orgID, _ := api.CurrentUserOrganizationID(apiClient, core.Hostname())
	if _, err := api.SendReport(apiClient, orgID, r); err != nil {
		log.Debugf("Error while sending report to reliably: %s", err)
	}

	opts.IO.StopProgressIndicator()

	for _, out := range opts.Outputs {

		var w io.Writer = os.Stdout
		if out.Path != "" {
			outfile, err := os.Create(out.Path) // creates or truncates with O_RDWR mode
			if err != nil {
				log.Error("error creating output file")
				log.Error(err)
				return err
			}
			w = outfile
			// we cannot defer outfile closing here as we are in a for-loop
		}
		report.Write(out.Format, r, w, opts.TemplateFile, log.StandardLogger(), &lr, &reports)

		if outfile, ok := w.(*os.File); ok {
			outfile.Close() // explicitly closing the file handle
		}

	}

	return nil
}

// watch - continuously fetch and update report
// and output to terminal
func watch(opts *ReportOptions) error {
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
			m, err := getManifest(opts.ManifestPath, opts.Service)
			if err != nil {
				log.Debug(err)
				errChan <- errors.New("an error occurred while attempting to load the manifest")
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
	var last *report.Report
	for {
		select {
		case r := <-rChan:
			clearScreen()
			fmt.Println(color.Magenta("Refreshing SLO report every 3 seconds."), "Press CTRL+C to quit.")
			report.Write(report.TABBED, r, os.Stdout, "", log.StandardLogger(), last, nil)
			last = r

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
func getManifest(manifestPath string, service string) (m *manifest.Manifest, err error) {

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

	return
}
