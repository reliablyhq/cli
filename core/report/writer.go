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

	case MARKDOWN:
		reportMarkdown(r, w)

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

func reportMarkdown(r *Report, w io.Writer) error {

	mdTemplate := `# SLO Report

Report time: {{ dateTime .Timestamp }}
{{ range $index, $service := .Services }}
## Service #({{ serviceNo $index}}): {{$service.Name}}

|  |  Type    | Name          | Actual | Target | Delta  |
|--| -------- | --------------| ------ | ------ | ------ | ---------------
{{ range $ind, $sl := $service.ServiceLevels }}{{ serviceLevelRow $sl }}
{{ end }}
{{ end }}

`

	t, err := template.New("slo-report").Funcs(markdownFuncMap()).Parse(mdTemplate)
	if err != nil {
		panic(err)
	}
	return t.Execute(w, r)
}

func getStatusIcon(res *ServiceLevelResult) string {
	if res == nil {
		return iconUnknown
	} else if res.sloIsMet {
		return iconTick
	} else {
		return iconEx

	}
}

func markdownFuncMap() template.FuncMap {
	// by default those functions return the given content untouched
	return template.FuncMap{
		"dateTime": func(t time.Time) string {
			return t.Format(time.RFC1123) + "  "
		},
		"bold": func(t string) string {
			return "**" + t + "**"
		},
		"serviceNo": func(i int) int {
			return i + 1
		},
		"serviceLevelRow": func(sl ServiceLevel) string {
			var builder strings.Builder
			statusIcon := getStatusIcon(sl.Result)
			unit := "%"
			fmt.Fprint(&builder, "|")
			fmt.Fprintf(&builder, "%s", statusIcon)
			fmt.Fprint(&builder, "|")
			fmt.Fprintf(&builder, "%s", sl.Type)
			fmt.Fprint(&builder, "|")
			fmt.Fprintf(&builder, "%s", sl.Name)
			fmt.Fprint(&builder, "|")
			fmt.Fprintf(&builder, "%.2f%s", sl.Result.Actual, unit)
			fmt.Fprint(&builder, "|")
			fmt.Fprintf(&builder, "%v%s", sl.Objective, unit)
			fmt.Fprint(&builder, "|")
			fmt.Fprintf(&builder, "%.2f%s", sl.Result.Delta, unit)
			fmt.Fprint(&builder, "|")
			period := sl.ObservationWindow.To.Sub(sl.ObservationWindow.From)
			fmt.Fprintf(&builder, "(last %s)", period)
			sl.ObservationWindow.To.Sub(sl.ObservationWindow.From)

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
