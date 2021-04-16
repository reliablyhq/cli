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

	return
}

func reportSimpleText(r *Report, w io.Writer) {

	for i, svc := range r.Services {
		fmt.Fprintf(w, color.Yellow(fmt.Sprintf("Service #%d: %s\n", i+1, svc.Name)))

		for _, sl := range svc.ServiceLevels {

			tick := iostreams.SuccessIcon()

			if sl.Result == nil {
				tick = iostreams.UnknownIcon()
				fmt.Fprintf(w, "%s %s\n", tick, sl.Name)

			} else {

				unit := "%"

				if sl.Type == "latency" {

					if float64(sl.Result.Actual.(float64)) > sl.Objective {
						tick = iostreams.FailureIcon()
					}
					unit = "ms"

				} else {
					if float64(sl.Result.Actual.(float64)) < sl.Objective {
						tick = iostreams.FailureIcon()
					}
				}

				fmt.Fprintf(w, "%s %s: %v%s (last %s) [objective: %v%s, delta: %v%s]\n",
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
		fmt.Println("service name is ", svc.Name)
		svcRowHeader := []string{color.Yellow(fmt.Sprintf("Service #%d: %s", i+1, svc.Name)), " ", " ", " "}
		table.Append(svcRowHeader)

		for _, sl := range svc.ServiceLevels {

			tick := iconTick
			unit := "%"
			colorFunc := color.Green

			if sl.Type == "latency" {
				unit = "ms"
			}

			if sl.Result == nil {
				tick = iconUnknown

				row := []string{
					fmt.Sprintf("%s %s", tick, sl.Name),
					"---",
					fmt.Sprintf("%v%s", sl.Objective, unit),
					"---"}
				table.Append(row)

			} else {

				if !sl.sloIsMet {
					tick = iconEx
					colorFunc = color.Red
				}

				row := []string{
					fmt.Sprintf("%s %s", tick, sl.Name),
					color.Bold(colorFunc(fmt.Sprintf("%v%s", sl.Result.Actual, unit))),
					fmt.Sprintf("%v%s", sl.Objective, unit),
					fmt.Sprintf("%v%s", sl.Result.Delta, unit)}
				table.Append(row)

			}

		}

		if i < len(r.Services)-1 {
			table.Append(emptyRow) // empty lines between services except last one
		}

	}

	/*
		var actual string
		var delta = func(actual, s string) string {
			if actual != "---" {
				return s
			}
			return actual
		}
	*/

	/*
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
	*/

	/*
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
	*/

	/*
		for _, row := range data {
			table.Append(fmt.Sprintf("# %s", sl.Name), "", "", "")
			table.Append(row)
			// add black row for spacing betwen services
			table.Append([]string{"", "", "", ""})
		}
	*/

	// render table
	table.Render()
}
