package report

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/reliablyhq/cli/core/errors"
	"github.com/reliablyhq/cli/core/manifest"
)

const (
	oneDay  = 24 * time.Hour
	oneWeek = 7 * oneDay
)

var (
	r = rand.New(rand.NewSource(time.Now().Unix()))
)

var notSuportedErrorBuilder = func(thingType, thingName string) error {
	return fmt.Errorf("%s type '%s' is not currently supported. I've informed ReliablyHQ about this; check back later - maybe we'll be able to help then.", thingType, thingName)
}

func GenerateReport(m *manifest.Manifest) (*Report, error) {
	r := Report{}
	allErrors := make([]error, 0)

	r.ApplicationName = m.ApplicationName
	r.Timestamp = time.Now().UTC()
	r.Targets.ErrorBudgetPercent = m.ServiceLevel.ErrorBudgetPercent
	r.Targets.ServiceLevel = m.ServiceLevel.Availability
	r.Targets.Latency = m.ServiceLevel.Latency
	r.Dependencies = m.Dependencies

	if actual, err := getCurrentErrorPc(m, oneWeek); err == nil {
		r.Delta.ErrorBudgetPercent = actual - m.ServiceLevel.ErrorBudgetPercent
	} else {
		allErrors = append(allErrors, err)
	}

	if actual, err := getCurrentAvailability(m, oneWeek); err == nil {
		r.Delta.ServiceLevel = actual - m.ServiceLevel.Availability
	} else {
		allErrors = append(allErrors, err)
	}

	if actual, err := get99PercentLatency(m, oneWeek); err == nil {
		r.Delta.Latency = actual - m.ServiceLevel.Latency
	} else {
		allErrors = append(allErrors, err)
	}

	if len(allErrors) > 0 {
		return &r, errors.NewCompoundError("multiple errors occured", allErrors)
	}

	return &r, nil
}
