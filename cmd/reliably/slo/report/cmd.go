package report

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/MakeNowJust/heredoc/v2"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/reliablyhq/cli/cmd/reliably/cmdutil"
	"github.com/reliablyhq/cli/core/color"
	"github.com/reliablyhq/cli/core/iostreams"
	"github.com/reliablyhq/cli/core/report"
)

type Choice = cmdutil.Choice

const defaultFormat = "table"

var (
	supportedFormats  = Choice{"json", "yaml", "text", "table", "markdown", "template"}
	deprecatedFormats = Choice{"simple", "tabbed"}
)

func NewCommand(runF func(*report.ReportOptions) error) *cobra.Command {
	opts := &report.ReportOptions{
		IO: iostreams.System(),
	}

	cmd := &cobra.Command{
		Use:   "report",
		Short: "generate the SLOs report",
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

				opts.Outputs = append(opts.Outputs, report.ReportOutput{
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

	cmd.Flags().StringVarP(&opts.Selector, "selector", "l", "", "objectives selector based on labels")
	cmd.Flags().StringVarP(&opts.ManifestPath, "manifest", "m", "", "the location of the manifest file")
	cmd.Flags().StringVarP(&opts.OutputPath, "output", "o", "", "where the report should be written to")
	cmd.Flags().StringVarP(&opts.TemplateFile, "template", "t", "", "the name of the template to use for the report output")
	cmd.Flags().StringVarP(&opts.OutputFormat, "format", "f", "table", fmt.Sprintf("specify the report format. Allowed Values: %v", supportedFormats))
	cmd.Flags().BoolVarP(&opts.WatchFlag, "watch", "w", false, "continuously watch for changes in report output")
	cmd.Flags().StringVar(&opts.Service, "service", "", "the name of the service")

	return cmd
}

func reportRun(opts *report.ReportOptions) error {
	// check for -w/--watch
	if opts.WatchFlag {
		return Watch(opts)
	}

	opts.IO.StartProgressIndicator()

	reports, err := report.GetReports(opts)
	if err != nil {
		return fmt.Errorf("reports error: %w", err)
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
		// utils.Reverse(reportCollection)
		report.Write(out.Format, reports[0], w, opts.TemplateFile, log.StandardLogger(), reports[1], report.EditReportSlice(reports))

		if outfile, ok := w.(*os.File); ok {
			outfile.Close() // explicitly closing the file handle
		}

	}

	return nil
}

func Watch(opts *report.ReportOptions) error {
	rChan := make(chan []*report.Report, 5)
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
			reports, err := report.GetReports(opts)
			if err != nil {
				errChan <- err
			}
			rChan <- reports
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
			report.ClearScreen()
			fmt.Println(color.Magenta("Refreshing SLO report every 3 seconds."), "Press CTRL+C to quit.")
			report.Write(report.TABBED, r[0], os.Stdout, opts.TemplateFile, log.StandardLogger(), r[1], report.EditReportSlice(r))

		case err := <-errChan:
			return err

		case <-done:
			return nil
		}
	}

}
