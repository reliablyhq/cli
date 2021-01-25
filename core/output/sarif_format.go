package output

import (
	//"fmt"
	"strings"

	"github.com/reliablyhq/cli/core"
)

type sarifLevel string

const (
	sarifNone    = sarifLevel("none")
	sarifNote    = sarifLevel("note")
	sarifWarning = sarifLevel("warning")
	sarifError   = sarifLevel("error")
)

type sarifProperties struct {
	Tags []string `json:"tags"`
}

type sarifRule struct {
	ID string `json:"id"`
	//Name             string           `json:"name"`
	ShortDescription *sarifMessage `json:"shortDescription"`
	//FullDescription  *sarifMessage    `json:"fullDescription"`
	//Help             *sarifMessage    `json:"help"`
	Properties *sarifProperties `json:"properties"`
}

type sarifArtifactLocation struct {
	URI string `json:"uri"`
}

type sarifRegion struct {
	StartLine   uint64 `json:"startLine"`
	StartColumn uint64 `json:"startColumn"`
	//EndColumn   uint64 `json:"endColumn"`
}

type sarifPhysicalLocation struct {
	ArtifactLocation *sarifArtifactLocation `json:"artifactLocation"`
	Region           *sarifRegion           `json:"region"`
}

type sarifLocation struct {
	PhysicalLocation *sarifPhysicalLocation `json:"physicalLocation"`
}

type sarifMessage struct {
	Text string `json:"text"`
}

type sarifResult struct {
	RuleID    string           `json:"ruleId"`
	RuleIndex int              `json:"ruleIndex"`
	Level     sarifLevel       `json:"level"`
	Message   *sarifMessage    `json:"message"`
	Locations []*sarifLocation `json:"locations"`
}

type sarifDriver struct {
	Name           string       `json:"name"`
	Version        string       `json:"version"`
	InformationURI string       `json:"informationUri"`
	Rules          []*sarifRule `json:"rules,omitempty"`
}

type sarifTool struct {
	Driver *sarifDriver `json:"driver"`
}

type sarifArtifact struct {
	Location *sarifArtifactLocation `json:"location"`
}

type sarifRun struct {
	Tool      *sarifTool       `json:"tool"`
	Results   []*sarifResult   `json:"results"`
	Artifacts []*sarifArtifact `json:"artifacts"`
}

type sarifReport struct {
	Schema  string      `json:"$schema"`
	Version string      `json:"version"`
	Runs    []*sarifRun `json:"runs"`
}

// buildSarifReport return SARIF report struct
func buildSarifReport() *sarifReport {
	return &sarifReport{
		Version: "2.1.0",
		Schema:  "https://raw.githubusercontent.com/oasis-tcs/sarif-spec/master/Schemata/sarif-schema-2.1.0.json",
		Runs:    []*sarifRun{},
	}
}

// buildSarifRule return SARIF rule field struct
func buildSarifRule(suggestion *core.Suggestion) *sarifRule {

	var description string = suggestion.Message // by default, rule Definition is the same as Violation message
	if suggestion.RuleDef != "" {
		description = suggestion.RuleDef
	}

	return &sarifRule{
		ID: suggestion.RuleID,
		//Name: suggestion.Message,
		ShortDescription: &sarifMessage{
			Text: description,
		},
		/*
			FullDescription: &sarifMessage{
				Text: description,
			},
		*/
		/*
			Help: &sarifMessage{
				Text: fmt.Sprintf("%s\nPlatform: %s\nKind: %s\n", description, suggestion.Platform, suggestion.Kind),
			},
		*/
		Properties: &sarifProperties{
			Tags: []string{suggestion.Platform, suggestion.Kind},
		},
	}
}

// buildSarifLocation return SARIF location struct
func buildSarifLocation(suggestion *core.Suggestion, rootPath string) (*sarifLocation, error) {
	var filePath string = suggestion.File

	line := uint64(suggestion.Line)
	col := uint64(suggestion.Col)

	if rootPath != "" && strings.HasPrefix(suggestion.File, rootPath) {
		filePath = strings.Replace(suggestion.File, rootPath+"/", "", 1)
	}

	location := &sarifLocation{
		PhysicalLocation: &sarifPhysicalLocation{
			ArtifactLocation: &sarifArtifactLocation{
				URI: filePath,
			},
			Region: &sarifRegion{
				StartLine:   line,
				StartColumn: col,
				//EndColumn:   col,
			},
		},
	}

	return location, nil
}
