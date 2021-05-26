package report

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"text/template"
	"time"

	consolesize "github.com/nathan-fiscaletti/consolesize-go"
	"github.com/olekukonko/tablewriter"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"

	"github.com/reliablyhq/cli/core"
	"github.com/reliablyhq/cli/core/color"
	"github.com/reliablyhq/cli/core/iostreams"
	"github.com/reliablyhq/cli/utils"
	v "github.com/reliablyhq/cli/version"
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
func Write(format Format, r *Report, w io.Writer, l *log.Logger, lr *Report, lrs *[]Report) {
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
		_ = reportMarkdown(r, w, lrs)

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

				fmt.Fprintf(w, "%s %s: %.2f%s [objective: %v%s / %s, delta: %.2f%s, type: %s]\n",
					tick, sl.Name,
					sl.Result.Actual, unit,
					sl.Objective, unit,
					core.HumanizeDurationShort(sl.ObservationWindow.To.Sub(sl.ObservationWindow.From)),
					sl.Result.Delta, unit, sl.Type)
			}
		}

		if i < len(r.Services)-1 {
			fmt.Fprintln(w) // empty lines between services except last one
		}

	}
}

func reportMarkdown(r *Report, w io.Writer, lrs *[]Report) error {

	// combines the report and report history for use in the markdown report
	type ReportData struct {
		Rep   *Report
		Lreps *[]Report
	}
	// create report data from report & lrs
	rd := ReportData{r, lrs}

	cwd, err := os.Getwd()
	templatePath := "/report-templates/"
	templateName := "slo-rep.tmpl"

	// t, err := template.New("SLOTemplate").Funcs(markdownFuncMap()).Parse(SLOTemplate)
	t, err := template.New(templateName).Funcs(markdownFuncMap()).ParseFiles(cwd + templatePath + templateName)
	if err != nil {
		panic(err)
	}

	err = t.Execute(w, rd)
	if err != nil {
		panic(err)
	}

	return err
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
		"reliablyVersion": func() string {
			return v.Version + " built on: " + v.Date
		},
		"dateTime": func(t time.Time) string {
			return t.Format(time.RFC1123) + "  "
		},
		"bold": func(t string) string {
			return "**" + t + "**"
		},
		"serviceNo": func(i int) int {
			return i + 1
		},
		"svcLevelGetStatusIcon": func(sl ServiceLevel) string {
			statusIcon := getStatusIcon(sl.Result)
			return statusIcon
		},
		"svcLevelGetName": func(sl ServiceLevel) string {
			return sl.Name
		},
		"svcLevelGetActualResult": func(sl ServiceLevel) string {
			var result string
			unit := "%"
			if sl.Result != nil {
				result = fmt.Sprintf("%.2f%s", sl.Result.Actual, unit)
			} else {
				result = "---"
			}
			return result
		},
		"svcLevelGetObjective": func(sl ServiceLevel) string {
			unit := "%"
			return fmt.Sprintf("%v%s", sl.Objective, unit)
		},
		"svcLevelGetTimeWindow": func(sl ServiceLevel) string {
			period := sl.ObservationWindow.To.Sub(sl.ObservationWindow.From)
			return core.HumanizeDuration(period)
		},
		"svcLevelGetType": func(sl ServiceLevel) string {
			return sl.Type
		},
		"svcLevelGetTrends": func(svcName string, sl ServiceLevel, lrs *[]Report) string {
			var trends string = "n/a"
			if lrs != nil && len(*lrs) > 0 {
				slosAreMet := GetSLOTrend(svcName, sl.Name, *lrs)
				ticks := trendToTicks(slosAreMet)
				trends = strings.Join(ticks, " ") // Using non-breaking space here !!!
			}
			return trends
		},
		"errBudgetPercentage": func(sl ServiceLevel) string {
			unit := "%"
			errBudget := ErrorBudgetAsPercentage(sl.Objective)
			return fmt.Sprintf("%.2f %s", errBudget, unit)
		},
		"errBudgetAllowedDownTime": func(sl ServiceLevel) string {

			errBudget := ErrorBudgetAsPercentage(sl.Objective)
			period := getObservationWindow(&sl)
			allowedDowntime := DowntimePerPeriod(errBudget, period)

			return fmt.Sprintf("%v", allowedDowntime)
		},
		"errBudgetConsumed": func(sl ServiceLevel) string {

			errBudget := ErrorBudgetAsPercentage(sl.Objective)
			period := getObservationWindow(&sl)
			allowedDowntime := DowntimePerPeriod(errBudget, period)
			consumed, _ := getConsumedRemain(sl, errBudget, allowedDowntime)

			return fmt.Sprintf("%s", consumed)
		},
		"errBudgetRemain": func(sl ServiceLevel) string {

			errBudget := ErrorBudgetAsPercentage(sl.Objective)
			period := getObservationWindow(&sl)
			allowedDowntime := DowntimePerPeriod(errBudget, period)
			_, remain := getConsumedRemain(sl, errBudget, allowedDowntime)

			return fmt.Sprintf("%s", remain)
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
	var optTrendHeader string = ""
	if lrs != nil && len(*lrs) > 0 {
		optTrendHeader = "Trend"
	}
	table.SetHeader([]string{
		"",                                     //SLO name
		color.Bold(color.Magenta(" ")),         // empty separator column
		color.Bold(color.Magenta("  Current")), // non-breaking spaces to align right
		color.Bold(color.Magenta("Objective")),
		color.Bold(color.Magenta("/ Time Window")),
		color.Bold(color.Magenta(" ")), // empty separator column
		color.Bold(color.Magenta("Type")),
		color.Bold(color.Magenta(" ")), // empty separator column
		color.Bold(color.Magenta(optTrendHeader)),
	})
	table.SetColumnAlignment([]int{
		tablewriter.ALIGN_LEFT,
		tablewriter.ALIGN_CENTER,
		tablewriter.ALIGN_RIGHT,
		tablewriter.ALIGN_RIGHT,
		tablewriter.ALIGN_LEFT,
		tablewriter.ALIGN_CENTER,
		tablewriter.ALIGN_LEFT,
		tablewriter.ALIGN_CENTER,
		tablewriter.ALIGN_LEFT,
	})

	emptyRow := []string{""}

	for i, svc := range r.Services {
		svcRowHeader := []string{fmt.Sprintf("Service #%d: %s", i+1, svc.Name)}
		table.Append(svcRowHeader)

		for _, sl := range svc.ServiceLevels {

			tick := iconTick
			unit := "%"
			colorFunc := color.Green
			//_ = tick // make it does not complain about not-used variable ! hack

			// we compute the period between the real observation window time stamp
			// if no observed, we use the user defined time period
			period := getObservationWindow(sl)

			tuncatedSLName := utils.TruncateString(sl.Name, 78) // 80 chars max with tick

			//errBudget := ErrorBudgetAsPercentage(sl.Objective)
			//allowedDowntime := DowntimePerPeriod(errBudget, period)

			if sl.Result == nil {
				tick = "?"
				row := []string{
					fmt.Sprintf("%s %s", tick, tuncatedSLName),
					"",
					"--- ",
					fmt.Sprintf("%v%s", sl.Objective, unit),
					fmt.Sprintf("/ %s", core.HumanizeDurationShort(period)),
					"",
					strings.Title(sl.Type),
				}

				table.Rich(
					row,
					[]tablewriter.Colors{
						{tablewriter.FgHiBlackColor},
						{},
						{tablewriter.FgHiBlackColor},
						{tablewriter.FgHiBlackColor},
						{tablewriter.FgHiBlackColor},
						{},
						{tablewriter.FgHiBlackColor},
					},
				)

			} else {

				if !sl.Result.SloIsMet {
					tick = iconEx
					colorFunc = color.Red
				}

				var mov string = " " // progression compared to last report
				if last != nil {
					lastReportResult := last.GetResult(svc.Name, sl.Name)
					if lastReportResult != nil {
						mov = sloMovement(*sl.Result, *lastReportResult)
					}

				}
				_ = mov // hack force not to complain about not-used var

				var trends string
				if lrs != nil && len(*lrs) > 0 {
					slosAreMet := GetSLOTrend(svc.Name, sl.Name, *lrs)
					ticks := trendToTicks(slosAreMet)
					trends = strings.Join(ticks, " ") // Using non-breaking space here !!!
				}

				/*
					deltaDowntime := DowntimePerPeriod(sl.Result.Delta.(float64), sl.Period)

					c, r := ComsumedRemainingBudget(float64(100)-sl.Result.Actual.(float64), errBudget)
					c2, r2 := ruleOfThreeDuration(c, allowedDowntime, 100), ruleOfThreeDuration(r, allowedDowntime, 100)
					consumedVsAllowed := c2 - allowedDowntime
					consumed := core.HumanizeDurationShort(c2)
					if consumedVsAllowed > 0 {
						consumed = fmt.Sprintf("%s (+%s)", consumed, core.HumanizeDurationShort(consumedVsAllowed))
					}
					consumedAsPercent := float64(100) - sl.Result.Actual.(float64)
					remained := fmt.Sprintf("%s", core.HumanizeDurationShort(r2))

					log.Debugf("%s (%.2f%s)", consumed, consumedAsPercent, unit)
					log.Debugf("%s (%.2f%s)", core.HumanizeDurationShort(allowedDowntime), errBudget, unit)
				*/

				row := []string{
					fmt.Sprintf("%s %s", tick, tuncatedSLName),
					"",
					color.Bold(colorFunc(fmt.Sprintf("%.2f%s", sl.Result.Actual, unit))),
					fmt.Sprintf("%v%s", sl.Objective, unit),
					fmt.Sprintf("/ %s", core.HumanizeDurationShort(period)),
					"",
					strings.Title(sl.Type),
					"",
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

func getConsumedRemain(sl ServiceLevel, errBudget float64, allowedDowntime time.Duration) (string, string) {

	c, r := ComsumedRemainingBudget(float64(100)-sl.Result.Actual.(float64), errBudget)

	c2, r2 := ruleOfThreeDuration(c, allowedDowntime, 100), ruleOfThreeDuration(r, allowedDowntime, 100)
	consumedVsAllowed := c2 - allowedDowntime
	consumed := core.HumanizeDurationShort(c2)
	if consumedVsAllowed > 0 {
		consumed = fmt.Sprintf("%s (+%s)", consumed, core.HumanizeDurationShort(consumedVsAllowed))
	}

	remain := fmt.Sprintf("%s", core.HumanizeDurationShort(r2))

	return consumed, remain
}

// compute the period between the real observation window time stamp
// if not observed, we use the user defined time period
func getObservationWindow(sl *ServiceLevel) time.Duration {
	period := sl.ObservationWindow.To.Sub(sl.ObservationWindow.From)
	if period == 0 {
		period = sl.Period.ToDuration()
	}
	return period
}

// trend to ticks is a utility function that iterate over trending
// for SLO met/unmet and returns a list of ticks accordingly
func trendToTicks(trend []bool) []string {
	var ticks []string = make([]string, 0)

	for _, t := range trend {
		switch t {
		case true:
			ticks = append(ticks, iostreams.SuccessIcon())
		case false:
			ticks = append(ticks, iostreams.FailureIcon())
		}

	}

	return ticks
}

// sloMovement returns the progression icon of the SLO current value
// compared to the value from the previous report
func sloMovement(current ServiceLevelResult, previous ServiceLevelResult) (mov string) {

	if _, ok := current.Actual.(float64); !ok {
		return
	}
	if _, ok := previous.Actual.(float64); !ok {
		return
	}

	switch diff := current.Actual.(float64) - previous.Actual.(float64); {
	case diff == 0:
		mov = "="
	case diff < 0:
		mov = "↓"
	case diff > 0:
		mov = "↑"
	default:
		mov = " " // non-breaking space by default, we don't want to have it stripped
	}
	return
}

// ErrorBudgetAsPercentage returns the error budget for a SLO as percentage
func ErrorBudgetAsPercentage(slo float64) float64 {
	return (1 - slo/100) * 100
}

// DowntimePerPeriod computes the duration of allowed downtime for a
// given SLO percentage over a time period
func DowntimePerPeriod(percent float64, period time.Duration) time.Duration {
	p := period.Milliseconds()
	d := int64(percent * float64(p))
	return time.Duration(d * 1000).Truncate(time.Second)
}

// ruleOfThreeDuration computes the rule of three of a duration related to a given percentage's duration
func ruleOfThreeDuration(percent float64, allowed time.Duration, targetPercent float64) time.Duration {
	d := percent * float64(allowed.Milliseconds()) / float64(targetPercent)
	return time.Duration(d * 1000 * 1000).Truncate(time.Second)
}

// ComsumedRemainingBudget computes the consumed & remaining error budget percentages
func ComsumedRemainingBudget(delta float64, errBudget float64) (float64, float64) {

	c := ufloat64(delta) * 100 / ufloat64(errBudget)

	var r float64
	if c < 100 {
		r = 100 - ufloat64(c)
	}

	return c, r
}

// ufloat64 returns an unsigned float64
func ufloat64(f float64) float64 {
	if f < 0 {
		return -f
	}

	return f
}
