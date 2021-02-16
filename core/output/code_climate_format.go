package output

import (
	"github.com/reliablyhq/cli/core"
)

type ccSeverity string

const (
	ccInfo     = ccSeverity("info")
	ccMinor    = ccSeverity("minor")
	ccMajor    = ccSeverity("major")
	ccCritical = ccSeverity("critical")
	ccBlocker  = ccSeverity("blocker")
)

type ccCategory string

const (
	ccBug           = ccCategory("Bug Risk")
	ccClarity       = ccCategory("Clarity")
	ccCompatibility = ccCategory("Compatibility")
	ccComplexity    = ccCategory("Complexity")
	ccDuplication   = ccCategory("Duplication")
	ccPerformance   = ccCategory("Performance")
	ccSecurity      = ccCategory("Security")
	ccStyle         = ccCategory("Style")
)

// ReportFormat enumerates the output format for reported issues
type ReportFormat int

const (
	// ReportText is the default format that writes to stdout
	ReportText ReportFormat = iota // Plain text format

	// ReportJSON set the output format to json
	ReportJSON // Json format

	// ReportCSV set the output format to csv
	ReportCSV // CSV format

	// ReportJUnitXML set the output format to junit xml
	ReportJUnitXML // JUnit XML format

	// ReportSARIF set the output format to SARIF
	ReportSARIF // SARIF format

	//SonarqubeEffortMinutes effort to fix in minutes
	SonarqubeEffortMinutes = 5
)

type ccIssue struct {
	Type        string `json:"type"`        // Required
	CheckName   string `json:"check_name"`  // Required
	Description string `json:"description"` // Required
	//Content            string           `json:"content"`  // Optional
	Categories  []ccCategory `json:"categories"`  // Required
	Location    *ccLocation  `json:"location"`    // Required
	Severity    ccSeverity   `json:"severity"`    // Optional
	Fingerprint string       `json:"fingerprint"` // Optional
}

type ccLocation struct {
	Path  string           `json:"path"`
	Lines *ccLocationLines `json:"lines"` // either Lines or Positions
	//Positions  *ccLocationPositions  `json:"positions"`
}

type ccLocationLines struct {
	Begin uint64 `json:"begin"`
	End   uint64 `json:"end"`
}

type ccLocationPositions struct {
	Begin ccLocationPosition `json:"begin"`
	End   ccLocationPosition `json:"end"`
}

type ccLocationPosition struct {
	Line   uint64 `json:"line"`
	Column uint64 `json:"column"`
}

// Code Climate report is directly a list of CC issues
type ccReport []*ccIssue

// buildCcIssue return CodeClimate issue struct
func buildCcIssue(suggestion *core.Suggestion) (*ccIssue, error) {
	location, _ := buildCcLocation(suggestion)
	issue := &ccIssue{
		Type:        "issue",
		CheckName:   suggestion.RuleID,
		Description: suggestion.Message,
		Categories:  []ccCategory{ccCategory("Reliability")},
		Location:    location,
		Severity:    levelToCCSecurity(suggestion.Level),
		// CAUTIOUS !! the fingerprint is based on the issue location
		// so it's not safe to be used !!
		Fingerprint: suggestion.Fingerprint(),
	}

	return issue, nil
}

// levelToCCSecurity maps the suggestion level to CC Severity
func levelToCCSecurity(l core.Level) ccSeverity {
	var s ccSeverity
	switch l {
	case core.Info:
		s = ccInfo
	case core.Warning:
		s = ccMinor
	case core.Error:
		s = ccMajor
	default:
		s = ccInfo
	}
	return s
}

// buildCcLocation return Code Climate location struct
func buildCcLocation(suggestion *core.Suggestion) (*ccLocation, error) {
	var filePath string = suggestion.File

	line := uint64(suggestion.Line)
	// col := uint64(suggestion.Col)

	location := &ccLocation{
		Path: filePath,
		Lines: &ccLocationLines{
			Begin: line,
			End:   line,
		},
	}

	return location, nil
}
