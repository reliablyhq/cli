package core

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
)

// Suggestion is returned by a policy if it discovers a violation with the scanned code.
type Suggestion struct {
	//Severity   Score  `json:"severity"`   // issue severity (how problematic it is)
	//Confidence Score  `json:"confidence"` // issue confidence (how sure we are we found it)
	RuleID  string `json:"rule_id"`         // Rule identifier
	RuleDef string `json:"rule_definition"` // Rule definition
	Message string `json:"details"`         // Human readable explanation
	Level   Level  `json:"level"`           // level
	File    string `json:"file"`            // File name we found it in
	//Code       string `json:"code"`       // Impacted code line
	Line    int    `json:"line"`       // Line number in file
	Col     int    `json:"column"`     // Column number in line
	Example string `json:"-" yaml:"-"` // Example of valid rule usage

	Platform string `json:"platform"` // Platform handling the resource
	Kind     string `json:"type"`     // Type of resource
	Name     string `json:"name"`     // Name of the resource

	Hash string `json:"-" yaml:"-"` // Unique Hash identifying the suggestion - not exported - used as fingerprint if specified
}

// UnmarshalJSON unmarshal json string into object
// by handling custom level string-to-int conversion
func (s *Suggestion) UnmarshalJSON(data []byte) error {

	type Alias Suggestion

	a := &struct {
		Level string `json:"level"`
		*Alias
	}{
		Alias: (*Alias)(s),
	}
	if err := json.Unmarshal(data, &a); err != nil {
		return err
	}

	if a.Level == "" {
		return nil
	}

	level, err := NewLevel(a.Level)
	if err != nil {
		return fmt.Errorf("cannot unmarshal string into Go struct field Suggestion.Level of type Level: %w", err)
	}

	s.Level = level

	return nil
}

// FileLocation point out the file path and line/column numbers in file
func (s Suggestion) FileLocation() string {
	return fmt.Sprintf("%s:%v:%v", s.File, s.Line, s.Col)
}

// Fingerprint generates a unqiue hash for the current suggestion
// based on unique context values, but not location.
// As better explained in the SARIF spec:
// This value shall be the same for results that are logically identical,
// and distinct for any two suggestions that are logically distinct.
// It must be resistant to changes that do not affect the logical identity
// of the result, such as location whithin a source file.
func (s Suggestion) Fingerprint() string {

	if s.Hash != "" {
		return s.Hash
	}

	raw := fmt.Sprintf("%s:%s:%s:%s:%s", s.File, s.Platform, s.Kind, s.Name, s.RuleID)
	hash := md5.Sum([]byte(raw))
	return fmt.Sprintf("%x", hash)
}

/*
// MarshalJSON is used convert a Score object into a JSON representation
func (c Score) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.String())
}
*/

/*
// String converts a Score into a string
func (c Score) String() string {
	switch c {
	case High:
		return "HIGH"
	case Medium:
		return "MEDIUM"
	case Low:
		return "LOW"
	}
	return "UNDEFINED"
}
*/

// NewSuggestion creates a new Suggestion
// It basically converts the inner nested structure into a simple one
// that holds all information needed for report formatting
func NewSuggestion(result Result, live bool) *Suggestion {

	var filePath string
	var row int
	var col int

	if live {
		filePath = result.Resource.Kind + ":" + result.Resource.Name
		row = -1
		col = -1
	} else {
		filePath = result.Resource.File.Filepath
		row = result.Location.Row
		col = result.Location.Col
	}
	return &Suggestion{
		File:     filePath,
		Line:     row,
		Col:      col,
		RuleID:   result.Rule.ID,
		RuleDef:  result.Rule.Definition,
		Level:    result.Rule.Level,
		Message:  result.Message,
		Platform: result.Resource.Platform,
		Kind:     result.Resource.Kind,
		Name:     result.Resource.Name,
		Example:  result.Example,
	}
}
