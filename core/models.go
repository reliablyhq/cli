package core

import (
	"bufio"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/reliablyhq/cli/core/color"
	//"fmt"
	//"io"
	//log "github.com/sirupsen/logrus"
)

// File represent a file location on file system
type File struct {
	Filepath string
}

// Resource is a resource to analyze from a file.
// A file can contain multiple resource, indicated by the startingLine
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
	Example  string
}

// ResultSet is a list of results after analysis
// It can contain results for multiple resources for multiple files
type ResultSet []Result

// SafeWriter allows to safely output to writer until an error occurs
type SafeWriter struct {
	w *bufio.Writer
	//w		io.Writer
	err error
}

// Writeln safewrite writes a string with the inner io writer
// If an error occurred on previous write, the next strings will be ignored
// The string will always be terminated by a line return char, ie. it will
// be written if the given string does not termiate with a CRLF
func (sw *SafeWriter) Writeln(s string) {

	if sw.err != nil {
		// does not write if an error already occurred
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

const (
	levelInfo    = "info"
	levelWarning = "warning"
	levelError   = "error"
)

func (l Level) String() string {
	var str string
	switch l {
	case Info:
		str = levelInfo
	case Warning:
		str = levelWarning
	case Error:
		str = levelError
	default:
		str = ""
	}
	return str
}

func (l Level) ColoredString() string {
	var str string
	switch l {
	case Info:
		str = color.Yellow(levelInfo)
	case Warning:
		str = color.Magenta(levelWarning)
	case Error:
		str = color.Red(levelError)
	default:
		str = ""
	}
	return str
}

// ColoredSquare is a function that will return a string with a
// colored square ("???"), where the color is determined by the level
func (l Level) ColoredSquare() string {
	var str string
	switch l {
	case Info:
		str = color.Yellow("???")
	case Warning:
		str = color.Magenta("???")
	case Error:
		str = color.Red("???")
	default:
		str = " "
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

// NewLevel returns a Level value from the matching string representation
func NewLevel(level string) (l Level, err error) {
	ll := strings.ToLower(level)
	switch ll {
	case levelInfo, levelWarning, levelError:
		l = LevelStringMap[level]
	default:
		err = fmt.Errorf("Invalid Level '%s'", level)
	}

	return
}

var LevelStringMap map[string]Level = map[string]Level{
	levelInfo:    Info,
	levelWarning: Warning,
	levelError:   Error,
}

var Levels []string = []string{levelInfo, levelWarning, levelError}
