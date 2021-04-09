package report

import (
	"encoding/json"
	"fmt"
	"io"

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

	default:
		tabbedoutput(r, w)
	}
	return
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
