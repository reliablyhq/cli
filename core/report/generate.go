package report

import (
	go_errors "errors"
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
	if m == nil {
		return nil, go_errors.New("manifest is nil")
	}

	if m.App == nil {
		return nil, go_errors.New("m.App is nil")
	}

	r := Report{}
	allErrors := make([]error, 0)

	r.ApplicationName = m.App.Name
	r.Timestamp = time.Now().UTC()

	if m.ServiceLevel != nil {
		r.Targets.ErrorBudgetPercent = m.ServiceLevel.ErrorBudgetPercent
		r.Targets.ServiceLevel = m.ServiceLevel.Availability
		r.Targets.LatencyMs = m.ServiceLevel.Latency.Milliseconds()

		if delta, err := getServiceLevelDeltas(m); err == nil {
			r.Delta = delta
		} else {
			allErrors = append(allErrors, err)
		}
	}

	r.Dependencies = m.Dependencies

	if len(allErrors) > 0 {
		return &r, errors.NewCompoundError("multiple errors occured", allErrors)
	}

	return &r, nil
}

func getServiceLevelDeltas(m *manifest.Manifest) (*ServiceLevelDelta, error) {
	if m == nil {
		return nil, go_errors.New("manifest is nil")
	}

	if m.ServiceLevel == nil {
		return nil, go_errors.New("manifest.ServiceLevel is nil")
	}

	var d ServiceLevelDelta
	errors := make([]error, 0)

	if actual, err := getCurrentErrorPc(m, oneWeek); err == nil {
		d.ErrorBudgetPercent = actual - m.ServiceLevel.ErrorBudgetPercent
	} else {
		errors = append(errors, err)
	}

	if actual, err := getCurrentAvailability(m, oneWeek); err == nil {
		d.ServiceLevel = actual - m.ServiceLevel.Availability
	} else {
		errors = append(errors, err)
	}

	if actual, err := get99PercentLatency(m, oneWeek); err == nil {
		d.LatencyMs = actual.Milliseconds() - m.ServiceLevel.Latency.Milliseconds()
	} else {
		errors = append(errors, err)
	}

	return &d, nil
}
