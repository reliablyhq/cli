package advice

import (
	"fmt"

	"github.com/reliablyhq/cli/core"
)

var notSuportedErrorBuilder = func(thingType, thingName string) error {
	return fmt.Errorf("%s type '%s' is not currently supported. I've informed ReliablyHQ about this, check back later - maybe we'll be able to help then.", thingType, thingName)
}

func GetAdviceFor(m *core.Manifest) ([]*Advice, error) {
	var err error
	allAdvice := make([]*Advice, 0)

	if a, err := getAdviceForType(m.Type); err != nil {
		return allAdvice, err
	} else {
		allAdvice = append(allAdvice, a...)
	}

	if a, err := getAdviceForCI(m.CI.Type); err != nil {
		return allAdvice, err
	} else {
		allAdvice = append(allAdvice, a...)
	}

	return allAdvice, err
}
