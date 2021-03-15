package advice

import (
	"fmt"

	"github.com/reliablyhq/cli/core"
	"github.com/reliablyhq/cli/core/errors"
)

type CompoundError struct {
	innerErrors []error
}

var notSuportedErrorBuilder = func(thingType, thingName string) error {
	return fmt.Errorf("%s type '%s' is not currently supported. I've informed ReliablyHQ about this; check back later - maybe we'll be able to help then.", thingType, thingName)
}

func GetAdviceFor(m *core.Manifest) ([]*Advice, error) {
	allAdvice := make([]*Advice, 0)
	allErrors := make([]error, 0)

	if a, err := getAdviceForType(m.Type); err == nil {
		allAdvice = append(allAdvice, a...)
	} else {
		allErrors = append(allErrors, err)
	}

	if a, err := getAdviceForCI(m.CI.Type); err == nil {
		allAdvice = append(allAdvice, a...)
	} else {
		allErrors = append(allErrors, err)
	}

	return allAdvice, errors.NewCompoundError("multiple errors occured", allErrors)
}
