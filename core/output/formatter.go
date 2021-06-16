package output

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"
	"text/tabwriter"
	plainTemplate "text/template"

	fColor "github.com/fatih/color"
	gColor "github.com/gookit/color"
	"github.com/nathan-fiscaletti/consolesize-go"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"

	"github.com/reliablyhq/cli/core"
	"github.com/reliablyhq/cli/core/color"
	"github.com/reliablyhq/cli/utils"
	"github.com/reliablyhq/cli/version"
)

const sortFlag = true

var text = `{{ notice "Results:" }}
{{ range $index, $issue := .Suggestions }}
{{ prompt ">" }} {{ bold $issue.FileLocation }} [{{ printLevel $issue.Level }}] {{ bold $issue.Message }}
Rule: {{ $issue.RuleID }}, Platform: {{ $issue.Platform}}, Kind: {{ $issue.Kind }}
{{if $issue.Example }}
# Example:
{{ $issue.Example}}
{{ end }}
{{ end }}
{{ notice "Summary:" }}
	{{ $count := len .Suggestions }}{{ if eq $count 0 }}
	{{- success "No suggestion found" }}
	{{- else }}{{ if eq $count 1 }}
			{{- danger $count " suggestion found" }}
		{{- else }}
			{{- danger $count " suggestions found" }}
		{{- end }}
	{{ counter "info" .Counters.info }} - {{ counter "warning" .Counters.warning }} - {{ counter "error" .Counters.error }}
	{{- end }}

`

var textTemplateTabbed = `Results:{{ range $index, $issue := .Suggestions }}
{{ levelSymbol $issue.Level}}	{{ printLiveFileLocation $issue.FileLocation $issue.Kind }}	{{ $issue.Platform}}:{{ $issue.Kind }}	{{ $issue.RuleID  }}	{{ $issue.Message }} {{ end }}
{{ notice "Summary:" }}
	{{ $count := len .Suggestions }}{{ if eq $count 0 }}
	{{- success "No suggestion found" }}
	{{- else }}{{ if eq $count 1 }}
			{{- danger $count " suggestion found" }}
		{{- else }}
			{{- danger $count " suggestions found" }}
		{{- end }}
	{{ counter "info" .Counters.info }} - {{ counter "warning" .Counters.warning }} - {{ counter "error" .Counters.error }}
	{{- end }}

	`

type reportInfo struct {
	Suggestions []*core.Suggestion `json:"suggestions"`
	Counters    map[string]int
}

type bySuggestion []*core.Suggestion

func countsPerLevel(suggestions []*core.Suggestion) map[string]int {

	var counters map[string]int = map[string]int{
		"info":    0,
		"warning": 0,
		"error":   0,
	}

	for _, s := range suggestions {
		l := s.Level.String()
		counters[l] = counters[l] + 1
	}

	return counters
}

// CreateReport generates a report for the reported suggestions into
// the specified format. The formats currently accepted are:
// json, yaml, sarif, basic and text.
func CreateReport(w io.Writer, format string, baseDir string, suggestions []*core.Suggestion) error {
	log.WithFields(log.Fields{
		"writer":      w,
		"format":      format,
		"baseDir":     baseDir,
		"suggestions": fmt.Sprintf("%v", len(suggestions)),
	}).Debug("CreateReport")

	data := &reportInfo{
		Suggestions: suggestions,
		Counters:    countsPerLevel(suggestions),
	}
	var err error
	switch format {
	case "json":
		err = reportJSON(w, data)
	case "yaml":
		err = reportYAML(w, data)
	case "extended":
		err = reportFromPlaintextTemplate(w, text, data)
	case "table", "tabbed":
		err = reportFromTabbedTextTemplate(w, data)
	case "text", "simple", "basic", "linter":
		err = reportLinter(w, data)
	case "sarif":
		err = reportSARIF(baseDir, w, data)
	case "codeclimate":
		err = reportCodeClimate(w, data)
	default:
		err = reportLinter(w, data)
	}
	return err
}

func reportJSON(w io.Writer, data *reportInfo) error {
	raw, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		return err
	}

	_, err = w.Write(raw)
	_, _ = io.WriteString(w, "\n")
	return err
}

func reportYAML(w io.Writer, data *reportInfo) error {
	raw, err := yaml.Marshal(data)
	if err != nil {
		return err
	}
	_, err = w.Write(raw)
	return err
}

func reportLinter(w io.Writer, data *reportInfo) error {
	// Output Format:
	// file:row:col: [extra] Human readable message (Other:Value, ...)
	// Output Sample:
	// manifest.yaml:1:1: Reliably namespace is forbidden (Platform: Kubernetes, Kind: Namespace)

	for _, s := range data.Suggestions {
		/*
			_, err := fmt.Fprintf(w, "%s %s (Platform: %s, Kind: %s, Name: %s)\n",
				s.FileLocation(),
				s.Message,
				s.Platform,
				s.Kind,
				s.Name,
			)
		*/
		_, err := fmt.Fprintf(w, "%s [%s] %s\n",
			s.FileLocation(),
			s.Level.ColoredString(),
			s.Message,
		)
		if err != nil {

			return err
		}
	}

	return nil
}

func reportSARIF(rootPath string, w io.Writer, data *reportInfo) error {
	sr, err := convertToSarifReport(rootPath, data)
	if err != nil {
		return err
	}
	raw, err := json.MarshalIndent(sr, "", "\t")
	if err != nil {
		return err
	}

	_, err = w.Write(raw)
	_, _ = io.WriteString(w, "\n")
	return err
}

func convertToSarifReport(rootPath string, data *reportInfo) (*sarifReport, error) {
	sr := buildSarifReport()

	var rules []*sarifRule
	var artifacts = make([]*sarifArtifact, 0, len(data.Suggestions)) // slice with default length 0, at max total length of Suggestions
	var results []*sarifResult

	// Map of files with suggestions being reported,
	// to easily get the list of unique file names (ie map keys)
	var files = make(map[string]struct{}) // struct{} takes no memory for value
	var seenRules = make(map[string]int)
	var rulesIndex = 0

	for _, suggestion := range data.Suggestions {
		//rules = append(rules, buildSarifRule(suggestion)) // This appends one rule per suggestion - duplicates -
		rule := buildSarifRule(suggestion)

		ruleID := rule.ID
		var index int
		var seen bool

		if ruleID != "" {
			// when we have a rule ID defined, we need to keep the same rule index
			// if rule has already be seen, we use its index, otherwise we
			// add it to the global list with the next rule index value available
			index, seen = seenRules[rule.ID]
			if !seen {
				seenRules[rule.ID] = rulesIndex
				index = seenRules[rule.ID]
				rulesIndex++
				rules = append(rules, rule)
			}
		} else {
			// when the rule has no ID, we adds it to the global list
			// with the next rule index value available
			index = rulesIndex
			rulesIndex++
			rules = append(rules, buildSarifRule(suggestion))
		}

		location, err := buildSarifLocation(suggestion, rootPath)
		if err != nil {
			return nil, err
		}

		// can only contain a single location
		var issueLocations [1]*sarifLocation = [1]*sarifLocation{location}

		// register the current suggestion location file to gobal map
		files[location.PhysicalLocation.ArtifactLocation.URI] = struct{}{}

		// Note from SARIF spec 3.27.16:
		// a SARIF producer SHOULD NOT populate the fingerprints property.
		result := &sarifResult{
			RuleID:    suggestion.RuleID,
			RuleIndex: index,
			Level:     levelToSarifLevel(suggestion.Level),
			Message: &sarifMessage{
				Text: suggestion.Message,
			},
			Locations: issueLocations[:],
		}

		results = append(results, result)
	}

	tool := &sarifTool{
		Driver: &sarifDriver{
			Name:           "reliably",
			Version:        version.Version,
			InformationURI: "https://github.com/reliablyhq/cli/",
			Rules:          rules,
		},
	}

	// iterate over the map of unique file names to generate the Artifacts
	for f := range files {
		artifacts = append(artifacts, &sarifArtifact{
			Location: &sarifArtifactLocation{
				URI: f,
			}})
	}

	run := &sarifRun{
		Tool:      tool,
		Results:   results,
		Artifacts: artifacts,
	}

	sr.Runs = append(sr.Runs, run)

	return sr, nil
}

// levelToSarifLevel returns the sarif level string related to current level
// ! we cannot return a sarifLevel type as it's not exported
func levelToSarifLevel(l core.Level) sarifLevel {
	var sl sarifLevel
	switch l {
	case core.Info:
		sl = sarifNone
	case core.Warning:
		sl = sarifWarning
	case core.Error:
		sl = sarifError
	default:
		sl = sarifNone
	}
	return sl
}

func reportFromPlaintextTemplate(w io.Writer, reportTemplate string, data *reportInfo) error {
	enableColor := !fColor.NoColor
	t, e := plainTemplate.
		New("reliably").
		Funcs(plainTextFuncMap(enableColor)).
		Parse(reportTemplate)
	if e != nil {
		return e
	}

	return t.Execute(w, data)
}

func (s bySuggestion) Len() int {
	return len(s)
}
func (s bySuggestion) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s bySuggestion) Less(i, j int) bool {
	return s[i].Level > s[j].Level
}

func reportFromTabbedTextTemplate(w io.Writer, data *reportInfo) error {

	padding := 2
	minWidth := 0
	tabWidth := 0
	const padChar = ' '

	reportBuffer := new(bytes.Buffer)

	tw := tabwriter.NewWriter(reportBuffer, minWidth, tabWidth, padding, padChar, 0)
	if sortFlag {
		sort.Sort(bySuggestion(data.Suggestions))
	}

	consoleColumns, _ := consolesize.GetConsoleSize()

	err := reportFromPlaintextTemplate(tw, textTemplateTabbed, data)

	if err != nil {

		return err
	}
	tw.Flush()

	reportString := reportBuffer.String()
	reportLines := strings.Split(reportString, "\n")
	for _, reportLine := range reportLines {
		fmt.Println(utils.TruncateString(reportLine, consoleColumns))
	}

	return nil
}

func plainTextFuncMap(enableColor bool) plainTemplate.FuncMap {
	if enableColor {
		return plainTemplate.FuncMap{
			"highlight": highlight,
			"bold":      gColor.Bold.Render,
			"danger":    gColor.Danger.Render,
			"notice":    gColor.Notice.Render,
			"success":   gColor.Success.Render,
			"prompt": func(s string) string {
				return gColor.Bold.Render(gColor.Yellow.Render(s))
			},
			"printCode": fmt.Sprint,
			"printLevel": func(l core.Level) string {
				return gColor.Bold.Render(l.ColoredString())
			},
			"printLiveFileLocation": printLiveFileLocation,
			"levelSymbol": func(l core.Level) string {
				return l.ColoredSquare()
			},
			"counter": func(level string, count int) string {
				var tick string
				switch level {
				case "info":
					tick = color.BgYellow(" ")
				case "warning":
					tick = color.BgMagenta(" ")
				case "error":
					tick = color.BgRed(" ")
				default:
					tick = ""
				}
				return fmt.Sprintf("%s %v %s", tick, count, level)
			},
		}
	}

	// by default those functions return the given content untouched
	return plainTemplate.FuncMap{
		"highlight": func(t string, i int) string {
			return t
		},
		"bold":      fmt.Sprint,
		"danger":    fmt.Sprint,
		"notice":    fmt.Sprint,
		"success":   fmt.Sprint,
		"prompt":    fmt.Sprint,
		"printCode": fmt.Sprint,
		"printLevel": func(l core.Level) string {
			return l.String()
		},
		"printLiveFileLocation": printLiveFileLocation,
		"levelSymbol": func(l core.Level) string {
			return l.String()[0:1]
		},
		"counter": func(level string, count int) string {
			return fmt.Sprintf("%v %s", count, level)
		},
	}
}

var (
	errorTheme   = gColor.New(gColor.FgLightWhite, gColor.BgRed)
	warningTheme = gColor.New(gColor.FgBlack, gColor.BgYellow)
	defaultTheme = gColor.New(gColor.FgWhite, gColor.BgBlack)
)

// highlight returns content t colored based on Score
func highlight(t string, s core.Score) string {
	switch s {
	case core.High:
		return errorTheme.Sprint(t)
	case core.Medium:
		return warningTheme.Sprint(t)
	default:
		return defaultTheme.Sprint(t)
	}
}

func reportCodeClimate(w io.Writer, data *reportInfo) error {
	ccr, err := convertToCcReport(data)
	if err != nil {
		return err
	}
	raw, err := json.MarshalIndent(ccr, "", "\t")
	if err != nil {
		return err
	}

	_, err = w.Write(raw)
	_, _ = io.WriteString(w, "\n")
	return err
}

func convertToCcReport(data *reportInfo) (*ccReport, error) {

	// from sarif_fromat.go: type ccReport []*ccIssue
	var issues ccReport

	for _, suggestion := range data.Suggestions {
		newIssue, _ := buildCcIssue(suggestion)
		issues = append(issues, newIssue)
	}

	return &issues, nil
}

func printLiveFileLocation(location string, kind string) string {
	oldLocation := location

	// Trim away the default location
	if strings.HasSuffix(oldLocation, ":1:1") {
		oldLocation = oldLocation[:len(oldLocation)-4]
	}

	r := kind + ":"
	return strings.Replace(oldLocation, r, "", 1)
}
