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

func FromManifest(m *manifest.Manifest) (*Report, error) {
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
		r.Actual = &ServiceLevelIndicators{
			ErrorBudgetPercent: getCurrentErrorPc(m, oneWeek),
			ServiceLevel:       getCurrentAvailability(m, oneWeek),
			LatencyMs:          get99PercentLatencyMs(m, oneWeek).Milliseconds(),
		}

		r.Target = &ServiceLevelIndicators{
			ErrorBudgetPercent: m.ServiceLevel.ErrorBudgetPercent,
			ServiceLevel:       m.ServiceLevel.Availability,
			LatencyMs:          m.ServiceLevel.Latency.Milliseconds(),
		}

		r.Delta = &ServiceLevelIndicators{
			ErrorBudgetPercent: r.Actual.ErrorBudgetPercent - r.Target.ErrorBudgetPercent,
			ServiceLevel:       r.Actual.ServiceLevel - r.Target.ServiceLevel,
			LatencyMs:          r.Actual.LatencyMs - r.Target.LatencyMs,
		}
	}

	r.Dependencies = m.Dependencies

	if len(allErrors) > 0 {
		return &r, errors.NewCompoundError("multiple errors occured", allErrors)
	}

	return &r, nil
}
