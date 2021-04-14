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

const (
	threshold                       = 0
	lessThan95pcAvailabilityMessage = "An availability of less than 95% allows more than 36.5 hours of downtime per month, which should be possible for any well built app deployed as a single instance. This availability target is probably not high enough for a production-ready system."
	errorBudgetExceededf            = "Your error budget has been exceeded by %.2f percent."
	errorBudgetTooLowf              = "You are under your error budget by %.2f percent. You could tighten your budget, or could decrease the quality of the experience your application provides (e.g by reducing the amount of resources given to your application)."
	latencyExceeded                 = "The average latency threshold has been exceeeded by %vms."
	latencyValid                    = "The average latency threshold has not been exceeded."
)

const (
	iconTick = "✅"
	iconEx   = "❌"
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

	if r.ServiceLevel.Delta == nil {
		l.Error("the report does not include a 'Delta'")
		return
	}

	switch format {
	case JSON:
		b, _ := json.MarshalIndent(r, "", "  ")
		fmt.Fprintln(w, string(b))

	case SimpleText:
		fmt.Printf("SLO report: (last %s)\n", r.ObservationWindow.To.Sub(r.ObservationWindow.From))

		if r.ServiceLevel.Delta.ErrorPercent > threshold {
			msg := fmt.Sprintf(errorBudgetExceededf, r.ServiceLevel.Delta.ErrorPercent)
			fmt.Printf("%s Error Rate: %.2f%%. %s\n", iostreams.FailureIcon(), r.ServiceLevel.Actual.ErrorPercent, msg)
		} else {
			msg := fmt.Sprintf(errorBudgetTooLowf, -r.ServiceLevel.Delta.ErrorPercent)
			fmt.Printf("%s Error Rate: %.2f%%. %s\n", iostreams.SuccessIcon(), r.ServiceLevel.Actual.ErrorPercent, msg)
		}

		if r.ServiceLevel.Delta.LatencyMs > threshold {
			msg := fmt.Sprintf(latencyExceeded, r.ServiceLevel.Delta.LatencyMs)
			fmt.Printf("%s Latency: %vms. %s\n", iostreams.FailureIcon(), r.ServiceLevel.Actual.LatencyMs, msg)
		} else {
			fmt.Printf("%s Latency: %vms. %s\n", iostreams.SuccessIcon(), r.ServiceLevel.Actual.LatencyMs, latencyValid)
		}

	case MARKDOWN:
		err := markdownOutput(r, w)
		if err != nil {
			l.Error("markdown report error %s", err)
		}
	default:
		tabbedoutput(r, w)
	}
	return
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


    Type            Actual    Target  Delta
--- ------------- --------   ------- ------ ---------------
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
			return t.Format(time.RFC1123) + "  "
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

			// fmt.Fprint(&builder, "|")
			fmt.Fprintf(&builder, "%s", statusIcon)
			fmt.Fprint(&builder, "    ")
			fmt.Fprint(&builder, "Error Rate")
			fmt.Fprint(&builder, "     ")
			fmt.Fprintf(&builder, "%.2f", sl.Actual.ErrorPercent)
			fmt.Fprint(&builder, "      ")
			fmt.Fprintf(&builder, "%.2f", sl.Target.ErrorPercent)
			fmt.Fprint(&builder, "   ")
			fmt.Fprintf(&builder, "%.2f%%", sl.Delta.ErrorPercent)
			fmt.Fprint(&builder, "   ")
			fmt.Fprintf(&builder, errorBudgetMsgF, sl.Delta.ErrorPercent)
			// fmt.Fprint(&builder, "|")

			return builder.String()
		},
		"latencyRow": func(sl ServiceLevel) string {
			var builder strings.Builder
			statusIcon := getStatusIcon(sl.Delta.LatencyMs > threshold)
			latencyMsg := getLatencyMsg(sl.Delta.LatencyMs > threshold, sl.Delta.LatencyMs)
			// fmt.Fprint(&builder, "|")
			fmt.Fprintf(&builder, "%s", statusIcon)
			fmt.Fprint(&builder, "    ")
			fmt.Fprint(&builder, "Latency")
			fmt.Fprint(&builder, "      ")
			fmt.Fprintf(&builder, "%dms", sl.Actual.LatencyMs)
			fmt.Fprint(&builder, "       ")
			fmt.Fprintf(&builder, "%dms", sl.Target.LatencyMs)
			fmt.Fprint(&builder, "    ")
			fmt.Fprintf(&builder, "%dms", sl.Delta.LatencyMs)
			fmt.Fprint(&builder, "    ")
			fmt.Fprint(&builder, latencyMsg)
			// fmt.Fprintln(&builder, "|")

			return builder.String()
		},
	}
}

func tabbedoutput(r *Report, w io.Writer) {
	fmt.Fprintf(w, "\n-----------\n[%s] SLO report: (last %s) \n-----------\n",
		color.Cyan(r.Name),
		r.ObservationWindow.To.Sub(r.ObservationWindow.From))

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
	// table.SetRowSeparator("--")
	table.SetColumnSeparator("")
	table.SetHeaderLine(false)
	table.SetColWidth(maxColWidth)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetHeader([]string{"",
		color.Bold(color.Magenta("Actual")),
		color.Bold(color.Magenta("Target")),
		color.Bold(color.Magenta("Delta")), ""})

	var data [][]string

	// set error budget data
	if r.ServiceLevel.Delta.ErrorPercent > threshold {
		data = append(data, []string{
			fmt.Sprintf("%s Error Rate", iconEx),
			color.Bold(color.Red(fmt.Sprintf("%.2f", r.ServiceLevel.Actual.ErrorPercent))),
			fmt.Sprintf("%.2f", r.ServiceLevel.Target.ErrorPercent),
			fmt.Sprintf("%.2f%%", r.ServiceLevel.Delta.ErrorPercent),
			fmt.Sprintf(errorBudgetExceededf, r.ServiceLevel.Delta.ErrorPercent),
		})
	} else {
		data = append(data, []string{
			fmt.Sprintf("%s Error Rate", iconTick),
			color.Bold(color.Green(fmt.Sprintf("%.2f", r.ServiceLevel.Actual.ErrorPercent))),
			fmt.Sprintf("%.2f", r.ServiceLevel.Target.ErrorPercent),
			fmt.Sprintf("%.2f%%", -r.ServiceLevel.Delta.ErrorPercent),
			fmt.Sprintf(errorBudgetTooLowf, -r.ServiceLevel.Delta.ErrorPercent),
		})
	}

	// set latency
	if r.ServiceLevel.Delta.LatencyMs > threshold {
		data = append(data, []string{
			fmt.Sprintf("%s Latency", iconEx),
			color.Bold(color.Red(fmt.Sprintf("%dms", r.ServiceLevel.Actual.LatencyMs))),
			fmt.Sprintf("%dms", r.ServiceLevel.Target.LatencyMs),
			fmt.Sprintf("%dms", r.ServiceLevel.Delta.LatencyMs),
			fmt.Sprintf(latencyExceeded, r.ServiceLevel.Delta.LatencyMs),
		})
	} else {
		data = append(data, []string{
			fmt.Sprintf("%s Latency", iconTick),
			color.Bold(color.Green(fmt.Sprintf("%dms", r.ServiceLevel.Actual.LatencyMs))),
			fmt.Sprintf("%dms", r.ServiceLevel.Target.LatencyMs),
			fmt.Sprintf("%dms", r.ServiceLevel.Delta.LatencyMs),
			latencyValid,
		})
	}

	for _, row := range data {
		table.Append(row)
		// add black row for spacing
		table.Append([]string{"", "", "", ""})
	}

	// render table
	table.Render()
}
