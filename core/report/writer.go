package report

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/olekukonko/tablewriter"
	"github.com/reliablyhq/cli/core/color"
	"github.com/sirupsen/logrus"
)

const (
	threshold                       = 0
	lessThan95pcAvailabilityMessage = "An availability of less than 95% allows more than 36.5 hours of downtime per month, which should be possible for any well built app deployed as a single instance. This availability target is probably not high enough for a production-ready system."
	errorBudgetExceededf            = "Your error budget has been exceeded by %.2f percent. This is pretty bad :("
	errorBudgetTooLowf              = "You are under your error budget by %.2f percent. You could tighten your budget, or could decrease the quality of the experience your application provides (e.g by reducing the amount of resources given to your application)."
	latencyExceeded                 = "The average latency threshold has been exceeeded by %vms"
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
		if r.ServiceLevel.Delta.ErrorPercent < threshold {
			l.Warnf(errorBudgetTooLowf, -r.ServiceLevel.Delta.ErrorPercent)
		} else if r.ServiceLevel.Delta.ErrorPercent > threshold {
			l.Warnf(errorBudgetExceededf, r.ServiceLevel.Delta.ErrorPercent)
		}

		if r.ServiceLevel.Delta.LatencyMs > threshold {
			l.Warnf(latencyExceeded, r.ServiceLevel.Delta.LatencyMs)
		}

	default:
		tabbedoutput(r, w)
	}
	return
}

func tabbedoutput(r *Report, w io.Writer) {
	fmt.Fprintf(w, "\n-----------\nSLO report: (%s) \n-----------\n",
		r.ObservationWindow.To.Sub(r.ObservationWindow.From))

	table := tablewriter.NewWriter(w)
	table.SetAutoFormatHeaders(false)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetBorder(false)
	table.SetRowLine(false)
	// table.SetRowSeparator("--")
	table.SetColumnSeparator("")
	table.SetHeaderLine(false)
	table.SetColWidth(100)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetHeader([]string{"",
		color.Bold(color.Magenta("Actual")),
		color.Bold(color.Magenta("Target")),
		color.Bold(color.Magenta("Delta")), ""})

	var data [][]string

	// set error budget data
	if r.ServiceLevel.Delta.ErrorPercent < threshold {
		data = append(data, []string{
			fmt.Sprintf("%s Error Budget", iconTick),
			fmt.Sprintf("%.2f", r.ServiceLevel.Actual.ErrorPercent),
			fmt.Sprintf("%.2f", r.ServiceLevel.Target.ErrorPercent),
			fmt.Sprintf("%.2f%%", -r.ServiceLevel.Delta.ErrorPercent),
			fmt.Sprintf(errorBudgetTooLowf, -r.ServiceLevel.Delta.ErrorPercent),
		})
	}

	if r.ServiceLevel.Delta.ErrorPercent > threshold {
		// l.Warnf(errorBudgetExceededf, r.ServiceLevel.Delta.ErrorPercent)
		data = append(data, []string{
			fmt.Sprintf("%s Error Budget", iconEx),
			color.Bold(color.Red(fmt.Sprintf("%.2f", r.ServiceLevel.Actual.ErrorPercent))),
			fmt.Sprintf("%.2f", r.ServiceLevel.Target.ErrorPercent),
			fmt.Sprintf("%.2f%%", r.ServiceLevel.Delta.ErrorPercent),
			fmt.Sprintf(errorBudgetExceededf, r.ServiceLevel.Delta.ErrorPercent),
		})
	}

	// set latency
	if r.ServiceLevel.Delta.LatencyMs > threshold {
		// l.Warnf(latencyExceeded, r.ServiceLevel.Delta.LatencyMs)
		data = append(data, []string{
			fmt.Sprintf("%s Latency", iconEx),
			color.Bold(color.Red(fmt.Sprintf("%dms", r.ServiceLevel.Actual.LatencyMs))),
			fmt.Sprintf("%dms", r.ServiceLevel.Target.LatencyMs),
			fmt.Sprintf("%dms", r.ServiceLevel.Delta.LatencyMs),
			fmt.Sprintf(latencyExceeded, r.ServiceLevel.Delta.LatencyMs),
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
