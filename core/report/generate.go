package report

import (
	go_errors "errors"
	"fmt"
	"time"

	"github.com/reliablyhq/cli/core/errors"
	"github.com/reliablyhq/cli/core/manifest"
	"github.com/reliablyhq/cli/core/metrics"
)

const (
	oneDay  = 24 * time.Hour
	oneWeek = 7 * oneDay
)

var (
	timestampFn func() time.Time = time.Now().UTC
)

var notSuportedErrorBuilder = func(thingType, thingName string) error {
	return fmt.Errorf("%s type '%s' is not currently supported. I've informed ReliablyHQ about this; check back later - maybe we'll be able to help then.", thingType, thingName)
}

func FromManifest(m *manifest.Manifest) (*Report, error) {
	if m == nil {
		return nil, go_errors.New("manifest is nil")
	}

	r := Report{}
	allErrors := make([]error, 0)

	r.ApplicationName = m.App.Name
	r.Timestamp = timestampFn()
	r.Dependencies = []string{}

	if m.Service != nil {
		allLatency := []float64{}
		allErrorPercentages := []float64{}

		if len(m.Service.Resources) == 0 {
			return nil, go_errors.New("you haven't told us about any resources, so we won't be able to give you a report. Sorry :(")
		}

		for _, resource := range m.Service.Resources {
			provider, err := getProviderForResource(resource.Provider)
			if err != nil {
				return nil, err
			}

			to := time.Now()
			from := to.Add(-oneDay)

			if l, err := provider.Get99PercentLatencyMetricForResource(resource.ID, from, to); err == nil {
				allLatency = append(allLatency, l)
			} else {
				return nil, err
			}

			if e, err := provider.GetErrorPercentageMetricForResource(resource.ID, from, to); err == nil {
				allErrorPercentages = append(allErrorPercentages, e)
			} else {
				return nil, err
			}
		}

		r.ServiceLevel = &ServiceLevel{
			Target: &ServiceLevelIndicators{
				ErrorPercent: m.Service.ErrorBudgetPercent,
				LatencyMs:    m.Service.Latency.Milliseconds(),
			},
			Actual: &ServiceLevelIndicators{
				ErrorPercent: average(allErrorPercentages),
				LatencyMs:    int64(average(allLatency)),
			},
			Delta: &ServiceLevelIndicators{},
		}

		r.ServiceLevel.Delta.ErrorPercent = r.ServiceLevel.Actual.ErrorPercent - r.ServiceLevel.Target.ErrorPercent
		r.ServiceLevel.Delta.LatencyMs = r.ServiceLevel.Actual.LatencyMs - r.ServiceLevel.Target.LatencyMs
	}

	if m.Dependencies != nil {
		r.Dependencies = m.Dependencies
	}

	if len(allErrors) > 0 {
		return &r, errors.NewCompoundError("multiple errors occured", allErrors)
	}

	return &r, nil
}

func getProviderForResource(providerID string) (metrics.Provider, error) {
	if factory, ok := metrics.ProviderFactories[providerID]; ok {
		return factory()
	}

	return nil, fmt.Errorf("No provider factory found for '%s'", providerID)
}
