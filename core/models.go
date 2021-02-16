package core

import (
	"bufio"
	"encoding/json"
	//"fmt"
	//"io"
	//log "github.com/sirupsen/logrus"
)

// File represent a file location on file system
type File struct {
	Filepath string
}

// Resource is a resource to analyze from a file.
// A file can contain mutliple resource, indicated by the startingLine
// The platform indicates the platform on which the resource belongs to
// The kind indicates the type of the resource
type Resource struct {
	File         File
	StartingLine int
	Platform     string
	Kind         string
	Name         string
	URI          string
}

// Rule contains the basic informations of a policy
type Rule struct {
	ID         string
	Definition string
	Level      Level
}

// Location indicates row and column numbers for a specific char in a file
type Location struct {
	Row int
	Col int
}

// Result is a result of evaluation for a given resource
type Result struct {
	Resource *Resource
	Location Location
	Rule     Rule
	Message  string
}

// ResultSet is a list of results after analysis
// It can contain results for mutliple resources for mutliple files
type ResultSet []Result

// SafeWriter allows to safely output to writer until an error occurs
type SafeWriter struct {
	w *bufio.Writer
	//w		io.Writer
	err error
}

// Writeln safewrite writes a string with the inner io writer
// If an error occured on previous write, the next strings will be ignored
// The string will always be terminated by a line return char, ie. it will
// be written if the given string does not termiate with a CRLF
func (sw *SafeWriter) Writeln(s string) {

	if sw.err != nil {
		// does not write if an error already occured
		return
	}
	//_, sw.err = fmt.Fprintln(sw.w, s)

	_, sw.err = sw.w.WriteString(s)

	// Ensure to have a carriage return
	last := s[len(s)-1:]
	if last != "\n" {
		_, sw.err = sw.w.WriteString("\n")
	}

	if sw.err == nil {
		sw.w.Flush()
	}

}

// NewSafeWriter is a constructor function to return `*SafeWriter`
func NewSafeWriter(w *bufio.Writer) *SafeWriter {
	//func NewSafeWriter(w io.Writer) *SafeWriter {
	return &SafeWriter{
		w: w,
	}
}

// Score type used by severity and confidence values
type Score int

const (
	// Low severity or confidence
	Low Score = iota
	// Medium severity or confidence
	Medium
	// High severity or confidence
	High
)

// Level type indicate the level of a suggestion
type Level int

const (
	// Information level (starts at 1)
	Info Level = iota + 1
	// Warning level
	Warning
	// Error level
	Error
)

func (l Level) String() string {
	var str string
	switch l {
	case Info:
		str = "info"
	case Warning:
		str = "warning"
	case Error:
		str = "error"
	default:
		str = ""
	}
	return str
}

// MarshalJSON is used convert a Level object into a JSON representation
func (l Level) MarshalJSON() ([]byte, error) {
	return json.Marshal(l.String())
}

// MarshalYAML is used convert a Level object into a YAML representation
func (l Level) MarshalYAML() (interface{}, error) {
	return l.String(), nil
}
