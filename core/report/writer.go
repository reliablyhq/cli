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
	"github.com/reliablyhq/cli/core"
	"github.com/reliablyhq/cli/core/color"
	"github.com/reliablyhq/cli/core/iostreams"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
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
	YAML       Format = "yaml"
)

// Write - write report based on given format
func Write(format Format, r *Report, w io.Writer, l *logrus.Logger, lr *Report, lrs *[]Report) {
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

	case YAML:
		b, _ := yaml.Marshal(r)
		fmt.Fprintln(w, string(b))

	case SimpleText:
		reportSimpleText(r, w)

	case MARKDOWN:
		_ = reportMarkdown(r, w)

	default:
		reportTable(r, w, lr, lrs)
	}

}

func reportSimpleText(r *Report, w io.Writer) {

	for i, svc := range r.Services {
		fmt.Fprint(w, color.Yellow(fmt.Sprintf("Service #%d: %s\n", i+1, svc.Name)))

		for _, sl := range svc.ServiceLevels {

			tick := iostreams.SuccessIcon()

			if sl.Result == nil {
				tick = iostreams.UnknownIcon()
				fmt.Fprintf(w, "%s %s\n", tick, color.Grey(sl.Name))

			} else {

				unit := "%"

				if !sl.Result.SloIsMet {
					tick = iostreams.FailureIcon()
				}

				fmt.Fprintf(w, "%s %s: %.2f%s (last %s) [objective: %v%s, delta: %.2f%s]\n",
					tick, sl.Name,
					sl.Result.Actual, unit,
					core.HumanizeDurationShort(sl.ObservationWindow.To.Sub(sl.ObservationWindow.From)),
					sl.Objective, unit,
					sl.Result.Delta, unit)

			}
		}

		if i < len(r.Services)-1 {
			fmt.Fprintln(w) // empty lines between services except last one
		}

	}
}

func reportMarkdown(r *Report, w io.Writer) error {

	mdTemplate := `# SLO Report

Report time: {{ dateTime .Timestamp }}
{{ range $index, $service := .Services }}
## Service #{{ serviceNo $index}}: {{$service.Name}}

|  |  Type    | Name          | Actual | Target | Delta  | Time Window  |
|--| -------- | --------------| ------ | ------ | ------ | ------------ |
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
	} else if res.SloIsMet {
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
			period := sl.ObservationWindow.To.Sub(sl.ObservationWindow.From)

			fmt.Fprint(&builder, "|")
			fmt.Fprintf(&builder, "%s", statusIcon)
			fmt.Fprint(&builder, "|")
			fmt.Fprintf(&builder, "%s", sl.Type)
			fmt.Fprint(&builder, "|")
			fmt.Fprintf(&builder, "%s", sl.Name)
			fmt.Fprint(&builder, "|")
			if sl.Result != nil {
				fmt.Fprintf(&builder, "%.2f%s", sl.Result.Actual, unit)
			} else {
				fmt.Fprint(&builder, "---")
			}
			fmt.Fprint(&builder, "|")
			fmt.Fprintf(&builder, "%v%s", sl.Objective, unit)
			fmt.Fprint(&builder, "|")
			if sl.Result != nil {
				fmt.Fprintf(&builder, "%.2f%s", sl.Result.Delta, unit)
			} else {
				fmt.Fprint(&builder, "---")
			}
			fmt.Fprint(&builder, "|")
			fmt.Fprint(&builder, core.HumanizeDuration(period))
			fmt.Fprint(&builder, "|")

			return builder.String()
		},
	}
}

func reportTable(r *Report, w io.Writer, last *Report, lrs *[]Report) {

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
	table.SetHeader([]string{
		"",
		color.Bold(color.Magenta("  Current")), // non-breaking spaces to align right
		color.Bold(color.Magenta("Target")),
		color.Bold(color.Magenta(" ")),       // empty separator column
		color.Bold(color.Magenta("  Delta")), // non-breaking spaces to align right
		//color.Bold(color.Magenta("Time Window")), // ! caution: we use non-breaking space to have header not on two lines !
		color.Bold(color.Magenta("Trend")),
	})
	table.SetColumnAlignment([]int{
		tablewriter.ALIGN_LEFT,
		tablewriter.ALIGN_RIGHT,
		tablewriter.ALIGN_LEFT,
		tablewriter.ALIGN_CENTER,
		tablewriter.ALIGN_RIGHT,
		tablewriter.ALIGN_LEFT,
	})

	emptyRow := []string{"", "", "", "", "", ""}

	for i, svc := range r.Services {
		svcRowHeader := []string{fmt.Sprintf("Service #%d: %s", i+1, svc.Name)}
		table.Append(svcRowHeader)

		// Using color breaks autowrap text to have weird behavior with unnecessary line return
		//svcRowHeader = []string{
		//  color.Yellow(fmt.Sprintf("Service #%d: %s", i+1, svc.Name))}
		//table.Append(svcRowHeader)

		for j, sl := range svc.ServiceLevels {

			tick := iconTick
			unit := "%"
			colorFunc := color.Green

			period := sl.ObservationWindow.To.Sub(sl.ObservationWindow.From)

			if sl.Result == nil {
				//tick = iconUnknown
				tick = "?"

				row := []string{
					fmt.Sprintf("%s %s", tick, sl.Name),
					"---",
					fmt.Sprintf("%v%s / %s", sl.Objective, unit, core.HumanizeDurationShort(period)),
					" ",
					"---",
					//core.HumanizeDuration(period),
					"",
				}

				table.Rich(row, []tablewriter.Colors{{tablewriter.FgHiBlackColor}, {tablewriter.FgHiBlackColor}, {tablewriter.FgHiBlackColor}, {}, {tablewriter.FgHiBlackColor}})
				//table.Append(row)

			} else {

				var lastReportResult *ServiceLevelResult
				// NB: current limitation, we find last/previous result at same
				// indexes than the current one
				if last != nil {
					lsvc := last.Services[i]
					if svc.Name == lsvc.Name {
						lsl := lsvc.ServiceLevels[j]
						if sl.Name == lsl.Name {
							lastReportResult = lsl.Result
						}
					}
				}
				// we need to be cautious with type assert here,
				// it could break entire rendering -> we need to check for type
				// ?? shall we round up to the same precision of the SLO objective ??
				var mov string
				if lastReportResult != nil {
					switch diff := sl.Result.Actual.(float64) - lastReportResult.Actual.(float64); {
					case diff == 0:
						mov = "="
						//trend = "←→"
					case diff < 0:
						// mov = color.Red("↓")
						mov = "↓"
					case diff > 0:
						// mov = color.Green("↑")
						mov = "↑"
					default:
						mov = " "
					}
				}

				var trends string
				if lrs != nil {

					l := len(*lrs)
					var slosAreMet = make([]string, l, l+1)

					for lastIndex, r := range *lrs {
						var wasMet string

						lsvc := r.Services[i]
						if svc.Name == lsvc.Name {
							lsl := lsvc.ServiceLevels[j]
							if sl.Name == lsl.Name {
								if lsl.Result != nil {
									switch lsl.Result.SloIsMet {
									case true:
										wasMet = iostreams.SuccessIcon()
									case false:
										wasMet = iostreams.FailureIcon()
									}
								}
							}
						}

						slosAreMet[l-lastIndex-1] = wasMet // in reversed order, from oldest to most recent
					}

					// we now append the current result SLO value to the trend
					if sl.Result.SloIsMet {
						slosAreMet = append(slosAreMet, iostreams.SuccessIcon())
					} else {
						slosAreMet = append(slosAreMet, iostreams.FailureIcon())
					}

					if !sl.Result.SloIsMet {
						tick = iconEx
						colorFunc = color.Red
					}

					trends = strings.Join(slosAreMet, " ") // Using non-breaking space here !!!
				}

				row := []string{
					fmt.Sprintf("%s %s", tick, sl.Name),
					fmt.Sprintf("%s %s", color.Bold(colorFunc(fmt.Sprintf("%.2f%s", sl.Result.Actual, unit))), mov),
					fmt.Sprintf("%v%s / %s", sl.Objective, unit, core.HumanizeDurationShort(period)),
					" ",
					fmt.Sprintf("%.2f%s", sl.Result.Delta, unit),
					//core.HumanizeDuration(period),
					trends,
				}
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
