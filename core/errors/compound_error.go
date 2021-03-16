package errors

import (
	"strings"
)

type CompoundError struct {
	message     string
	innerErrors []error
}

func NewCompoundError(message string, innerErrors []error) *CompoundError {
	return &CompoundError{message, innerErrors}
}

func (e CompoundError) Error() string {
	var sb strings.Builder

	var errorWriter func(*strings.Builder, string, error)
	errorWriter = func(sb *strings.Builder, prefix string, e error) {
		if x, ok := e.(CompoundError); ok {
			for _, innerE := range x.innerErrors {
				errorWriter(sb, prefix+"  ", innerE)
			}
		} else {
			sb.WriteString(prefix)
			sb.WriteString(e.Error())
			sb.WriteString("\n")
		}
	}

	sb.WriteString(e.message)
	sb.WriteString("\n")

	errorWriter(&sb, "", e)

	return sb.String()
}
