package report

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"text/template"
	"time"

	consolesize "github.com/nathan-fiscaletti/consolesize-go"
	"github.com/olekukonko/tablewriter"
	"github.com/reliablyhq/cli/core/color"
	"github.com/reliablyhq/cli/core/iostreams"
	"github.com/sirupsen/logrus"
)

/*
const (
	threshold                       = 0
	lessThan95pcAvailabilityMessage = "An availability of less than 95% allows more than 36.5 hours of downtime per month, which should be possible for any well built app deployed as a single instance. This availability target is probably not high enough for a production-ready system."
	errorBudgetExceededf            = "Your error budget has been exceeded by %.2f percent."
	errorBudgetTooLowf              = "You are under your error budget by %.2f percent. You could tighten your budget, or could decrease the quality of the experience your application provides (e.g by reducing the amount of resources given to your application)."
	latencyExceeded                 = "The average latency threshold has been exceeeded by %vms."
	latencyValid                    = "The average latency threshold has not been exceeded."
)
*/

const (
	iconTick    = "✅"
	iconEx      = "❌"
	iconUnknown = "❔"
)

// Format - type used to set supported report formats
type Format string

// Define supported formats
const (
	JSON       Format = "json"
	TABBED     Format = "tabbed"
	SimpleText Format = "simple"
	MARKDOWN   Format = "markdown"
)

// Write - write report based on given format
func Write(format Format, r *Report, w io.Writer, l *logrus.Logger) {
	if r == nil {
		return
	}

	if l == nil {
		return
	}

	switch format {
	case JSON:
		b, _ := json.MarshalIndent(r, "", "  ")
		fmt.Fprintln(w, string(b))

	case SimpleText:
		reportSimpleText(r, w)

	default:
		tabbedoutput(r, w)
	}

}

func reportSimpleText(r *Report, w io.Writer) {

	for i, svc := range r.Services {
		fmt.Fprint(w, color.Yellow(fmt.Sprintf("Service #%d: %s\n", i+1, svc.Name)))

		for _, sl := range svc.ServiceLevels {

			tick := iostreams.SuccessIcon()

			if sl.Result == nil {
				tick = iostreams.UnknownIcon()
				fmt.Fprintf(w, "%s %s\n", tick, sl.Name)

			} else {

				unit := "%"

				if !sl.Result.sloIsMet {
					tick = iostreams.FailureIcon()
				}

				fmt.Fprintf(w, "%s %s: %.2f%s (last %s) [objective: %v%s, delta: %.2f%s]\n",
					tick, sl.Name,
					sl.Result.Actual, unit,
					sl.ObservationWindow.To.Sub(sl.ObservationWindow.From),
					sl.Objective, unit,
					sl.Result.Delta, unit)

			}
		}

		if i < len(r.Services)-1 {
			fmt.Println() // empty lines between services except last one
		}

	}
}

type Todo struct {
	Name        string
	Description string
}

func markdownOutput(r *Report, w io.Writer) error {

	mdTemplate := `# SLO Report

SLO Name: {{ bold .Name }}

## Report Dates
Report time: {{ dateTime .Timestamp }}
Window Start time: {{ dateTime .ObservationWindow.From }}
Window End time: {{ dateTime .ObservationWindow.To }}
Window Duration: {{ duration .ObservationWindow.From .ObservationWindow.To }}

## SLO Summary

| |Type| Actual | Target | Delta | |
|----|----|---|----|---- |----|
{{ errorBudgetRow .ServiceLevel }}
{{ latencyRow .ServiceLevel }}
`

	t, err := template.New("slo-report").Funcs(markdownFuncMap()).Parse(mdTemplate)
	if err != nil {
		panic(err)
	}
	return t.Execute(w, r)
}

func getStatusIcon(e bool) string {
	if e {
		return iconEx
	} else {
		return iconTick
	}
}

func getErrorBudgetMsgF(e bool) string {
	if e {
		return errorBudgetExceededf
	} else {
		return errorBudgetTooLowf
	}
}

func getLatencyMsg(e bool, l int64) string {
	if e {
		return fmt.Sprintf(latencyExceeded, l)
	} else {
		return latencyValid
	}
}

func markdownFuncMap() template.FuncMap {
	// by default those functions return the given content untouched
	return template.FuncMap{
		"dateTime": func(t time.Time) string {
			return t.Format(time.RFC1123)
		},
		"duration": func(from time.Time, to time.Time) time.Duration {
			return to.Sub(from)
		},
		"bold": func(t string) string {
			return "**" + t + "**"
		},
		"errorRate": func(t string) string {
			return "**" + t + "**"
		},
		"errorBudgetRow": func(sl ServiceLevel) string {
			var builder strings.Builder
			statusIcon := getStatusIcon(sl.Delta.ErrorPercent > threshold)
			errorBudgetMsgF := getErrorBudgetMsgF(sl.Delta.ErrorPercent > threshold)

			fmt.Fprint(&builder, "|")
			fmt.Fprintf(&builder, "%s", statusIcon)
			fmt.Fprint(&builder, "|")
			fmt.Fprint(&builder, "Error Rate")
			fmt.Fprint(&builder, "|")
			fmt.Fprintf(&builder, "%.2f", sl.Actual.ErrorPercent)
			fmt.Fprint(&builder, "|")
			fmt.Fprintf(&builder, "%.2f", sl.Target.ErrorPercent)
			fmt.Fprint(&builder, "|")
			fmt.Fprintf(&builder, "%.2f%%", sl.Delta.ErrorPercent)
			fmt.Fprint(&builder, "|")
			fmt.Fprintf(&builder, errorBudgetMsgF, sl.Delta.ErrorPercent)
			fmt.Fprint(&builder, "|")

			return builder.String()
		},
		"latencyRow": func(sl ServiceLevel) string {
			var builder strings.Builder
			statusIcon := getStatusIcon(sl.Delta.LatencyMs > threshold)
			latencyMsg := getLatencyMsg(sl.Delta.LatencyMs > threshold, sl.Delta.LatencyMs)
			fmt.Fprint(&builder, "|")
			fmt.Fprintf(&builder, "%s", statusIcon)
			fmt.Fprint(&builder, "|")
			fmt.Fprint(&builder, "Latency")
			fmt.Fprint(&builder, "|")
			fmt.Fprintf(&builder, "%dms", sl.Actual.LatencyMs)
			fmt.Fprint(&builder, "|")
			fmt.Fprintf(&builder, "%dms", sl.Target.LatencyMs)
			fmt.Fprint(&builder, "|")
			fmt.Fprintf(&builder, "%dms", sl.Delta.LatencyMs)
			fmt.Fprint(&builder, "|")
			fmt.Fprint(&builder, latencyMsg)
			fmt.Fprintln(&builder, "|")

			return builder.String()
		},
	}
}

func tabbedoutput(r *Report, w io.Writer) {

	cols, _ := consolesize.GetConsoleSize()

	// compute max with for latest column that contains text on wrapped multi lines
	maxColWidth := cols - 44 // arbitrary based on error rate line length for first 3 cols
	if maxColWidth < 30 {
		maxColWidth = 30 // make it a default decent size
	}

	table := tablewriter.NewWriter(w)
	table.SetAutoFormatHeaders(false)
	table.SetAutoWrapText(true)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetBorder(false)
	table.SetRowLine(false)
	table.SetRowSeparator("")
	table.SetColumnSeparator("")
	table.SetHeaderLine(false)
	table.SetColWidth(maxColWidth)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetHeader([]string{"",
		color.Bold(color.Magenta("Actual")),
		color.Bold(color.Magenta("Target")),
		color.Bold(color.Magenta("Delta")), ""})

	emptyRow := []string{"", "", "", ""}

	for i, svc := range r.Services {
		svcRowHeader := []string{fmt.Sprintf("Service #%d: %s", i+1, svc.Name)}
		table.Append(svcRowHeader)

		// Using color breaks autowrap text to have weird behavior with unnecessary line return
		//svcRowHeader = []string{
		//  color.Yellow(fmt.Sprintf("Service #%d: %s", i+1, svc.Name))}
		//table.Append(svcRowHeader)

		for _, sl := range svc.ServiceLevels {

			tick := iconTick
			unit := "%"
			colorFunc := color.Green

			if sl.Result == nil {
				tick = iconUnknown

				row := []string{
					fmt.Sprintf("%s %s", tick, sl.Name),
					"---",
					fmt.Sprintf("%v%s", sl.Objective, unit),
					"---"}
				table.Append(row)

			} else {

				if !sl.Result.sloIsMet {
					tick = iconEx
					colorFunc = color.Red
				}

				row := []string{
					fmt.Sprintf("%s %s", tick, sl.Name),
					color.Bold(colorFunc(fmt.Sprintf("%.2f%s", sl.Result.Actual, unit))),
					fmt.Sprintf("%v%s", sl.Objective, unit),
					fmt.Sprintf("%.2f%s", sl.Result.Delta, unit)}
				table.Append(row)

			}
		}

		if i < len(r.Services)-1 {
			table.Append(emptyRow) // empty lines between services except last one
		}

	}

	// render table
	table.Render()
}
