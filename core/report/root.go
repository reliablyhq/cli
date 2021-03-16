package report

import (
	"fmt"

	"github.com/reliablyhq/cli/errors"
	"github.com/reliablyhq/cli/manifest"
)

var notSuportedErrorBuilder = func(thingType, thingName string) error {
	return fmt.Errorf("%s type '%s' is not currently supported. I've informed ReliablyHQ about this; check back later - maybe we'll be able to help then.", thingType, thingName)
}

func GetAdviceFor(m *manifest.Manifest) (*Advice, error) {
	advice := Advice{}
	allErrors := make([]error, 0)

	if m.ServiceLevel == nil {
		advice.Suggestions = append(advice.Suggestions, "I don't know about the service you are trying to provide. Run `reliably init service` to add a 'Service' block to your manifest.")
	} else {
		if s, err := getSuggestionsForService(m); err == nil {
			advice.Suggestions = append(advice.Suggestions, s...)
		} else {
			allErrors = append(allErrors, err)
		}
	}

	if m.CI == nil {
		advice.Suggestions = append(advice.Suggestions, "I don't know about the CI you are using. Run `reliably init ci` to add a 'CI' block to your manifest.")
	} else {
		if s, err := getSuggestionsForCI(m.CI); err == nil {
			advice.Suggestions = append(advice.Suggestions, s...)
		} else {
			allErrors = append(allErrors, err)
		}
	}

	if len(m.Apps) == 0 {
		advice.Suggestions = append(advice.Suggestions, "I don't know about the apps you are building. Run `reliably init apps` to add apps to your manifest.")
	} else {
		for _, app := range m.Apps {
			if s, err := getSuggestionsForApp(app); err == nil {
				advice.Suggestions = append(advice.Suggestions, s...)
			} else {
				allErrors = append(allErrors, err)
			}
		}
	}

	if len(allErrors) > 0 {
		return &advice, errors.NewCompoundError("multiple errors occured", allErrors)
	}

	return &advice, nil
}
