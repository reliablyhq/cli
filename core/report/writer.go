package report

import (
	"encoding/json"
	"fmt"
	"io"

	//consolesize "github.com/nathan-fiscaletti/consolesize-go"
	//"github.com/olekukonko/tablewriter"
	//"github.com/reliablyhq/cli/core/color"
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

	switch format {
	case JSON:
		b, _ := json.MarshalIndent(r, "", "  ")
		fmt.Fprintln(w, string(b))

	case SimpleText:
		reportSimpleText(r, w)

	default:
		return
		//tabbedoutput(r, w)

	}

	return
}

func reportSimpleText(r *Report, w io.Writer) {

	for i, svc := range r.Services {
		fmt.Fprintf(w, "# Service: %s\n", svc.Name)

		for _, sl := range svc.ServiceLevels {

			tick := iostreams.SuccessIcon()

			/*
				if sl.Type == "latency" {

				} else {

				}
			*/

			if sl.Result == nil {
				tick = iostreams.UnknownIcon()
				fmt.Fprintf(w, "%s %s\n", tick, sl.Name)

			} else {

				unit := "%"

				if sl.Type == "latency" {

					if float64(sl.Result.Actual.(float64)) > float64(sl.Result.Objective.(float64)) {
						tick = iostreams.FailureIcon()
					}
					unit = "ms"

				} else {
					if float64(sl.Result.Actual.(float64)) < float64(sl.Result.Objective.(float64)) {
						tick = iostreams.FailureIcon()
					}
				}

				fmt.Fprintf(w, "%s %s: %v%s (last %s) [objective: %v%s, delta: %v%s]\n",
					tick, sl.Name, sl.Result.Actual, unit,
					sl.ObservationWindow.To.Sub(sl.ObservationWindow.From),
					sl.Result.Objective, unit, sl.Result.Delta, unit)

			}

		}

		if i < len(r.Services)-1 {
			fmt.Println() // empty lines between services except last one
		}

	}

	/*
		fmt.Printf("SLO report: (last %s)\n", r.ObservationWindow.To.Sub(r.ObservationWindow.From))
		if !r.ServiceLevel.Actual.hasErrors(errPercentErr) {
			if r.ServiceLevel.Delta.ErrorPercent > threshold {
				msg := fmt.Sprintf(errorBudgetExceededf, r.ServiceLevel.Delta.ErrorPercent)
				fmt.Printf("%s Error Rate: %.2f%%. %s\n", iostreams.FailureIcon(), r.ServiceLevel.Actual.ErrorPercent, msg)
			} else {
				msg := fmt.Sprintf(errorBudgetTooLowf, -r.ServiceLevel.Delta.ErrorPercent)
				fmt.Printf("%s Error Rate: %.2f%%. %s\n", iostreams.SuccessIcon(), r.ServiceLevel.Actual.ErrorPercent, msg)
			}
		}

		if !r.ServiceLevel.Actual.hasErrors(latencyErr) {
			if r.ServiceLevel.Delta.LatencyMs > threshold {
				msg := fmt.Sprintf(latencyExceeded, r.ServiceLevel.Delta.LatencyMs)
				fmt.Printf("%s Latency: %vms. %s\n", iostreams.FailureIcon(), r.ServiceLevel.Actual.LatencyMs, msg)
			} else {
				fmt.Printf("%s Latency: %vms. %s\n", iostreams.SuccessIcon(), r.ServiceLevel.Actual.LatencyMs, latencyValid)
			}
		}

	*/

}

/*
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
	var actual string
	var delta = func(actual, s string) string {
		if actual != "---" {
			return s
		}
		return actual
	}

	// set error budget data
	actual = r.ServiceLevel.Actual.errorPercentString()
	if r.ServiceLevel.Delta.ErrorPercent > threshold {
		data = append(data, []string{
			fmt.Sprintf("%s Error Rate", iconEx),
			color.Bold(color.IfTrueRed(actual != "---", actual)),
			fmt.Sprintf("%.2f", r.ServiceLevel.Target.ErrorPercent),
			delta(actual, fmt.Sprintf("%.2f%%", r.ServiceLevel.Delta.ErrorPercent)),
		})
	} else {
		data = append(data, []string{
			fmt.Sprintf("%s Error Rate", iconTick),
			color.Bold(color.IfTrueGreen(actual != "---", actual)),
			fmt.Sprintf("%.2f", r.ServiceLevel.Target.ErrorPercent),
			delta(actual, fmt.Sprintf("%.2f%%", -r.ServiceLevel.Delta.ErrorPercent)),
		})
	}

	// set latency
	actual = r.ServiceLevel.Actual.latencyMsString()
	if r.ServiceLevel.Delta.LatencyMs > threshold {
		data = append(data, []string{
			fmt.Sprintf("%s Latency", iconEx),
			color.Bold(color.IfTrueRed(actual != "---", actual)),
			fmt.Sprintf("%dms", r.ServiceLevel.Target.LatencyMs),
			delta(actual, fmt.Sprintf("%dms", r.ServiceLevel.Delta.LatencyMs)),
		})
	} else {
		data = append(data, []string{
			fmt.Sprintf("%s Latency", iconTick),
			color.Bold(color.IfTrueGreen(actual != "---", actual)),
			fmt.Sprintf("%dms", r.ServiceLevel.Target.LatencyMs),
			delta(actual, fmt.Sprintf("%dms", r.ServiceLevel.Delta.LatencyMs)),
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
*/
