package report

import (
	go_errors "errors"
	"fmt"
	"time"

	"github.com/reliablyhq/cli/core/errors"
	"github.com/reliablyhq/cli/core/manifest"
	"github.com/reliablyhq/cli/core/metrics"
	log "github.com/sirupsen/logrus"
)

const (
	oneDay     = 24 * time.Hour
	oneWeek    = 7 * oneDay
	apiVersion = "1.0rc"
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

	r.APIVersion = apiVersion
	r.Timestamp = timestampFn()
	r.Dependencies = []string{}

	allLatency := []float64{}
	allErrorPercentages := []float64{}

	if m.Service == nil {
		return nil, go_errors.New("I don't know anything about your service objectives. Run `reliably slo init` to tell me.")
	}

	if len(m.Service.Resources) == 0 {
		return nil, go_errors.New("you haven't told us about any resources, so we won't be able to give you a report. Sorry :(")
	}

	for _, resource := range m.Service.Resources {
		provider, err := getProviderForResource(resource.Provider)
		if err != nil {
			log.Warnf("an error occured while getting a provider for resource: %s", resource.Provider)
			continue
		}

		to := time.Now()
		from := to.Add(-oneDay)

		if l, err := provider.Get99PercentLatencyMetricForResource(resource.ID, from, to); err == nil {
			allLatency = append(allLatency, l)
		} else {
			log.Warnf("an error occured while getting latency data for resource: %s-%s => %v ", resource.Provider, resource.ID, err)
		}

		if e, err := provider.GetErrorPercentageMetricForResource(resource.ID, from, to); err == nil {
			allErrorPercentages = append(allErrorPercentages, e)
		} else {
			log.Warnf("an error occured while getting error percentage data for resource: %s-%s => %v ", resource.Provider, resource.ID, err)
		}

	}

	r.ServiceLevel = &ServiceLevel{
		Target: &ServiceLevelIndicators{
			ErrorPercent: m.Service.Objective.ErrorBudgetPercent,
			LatencyMs:    m.Service.Objective.Latency.Milliseconds(),
		},
		Actual: &ServiceLevelIndicators{
			ErrorPercent: average(allErrorPercentages),
			LatencyMs:    int64(average(allLatency)),
		},
		Delta: &ServiceLevelIndicators{},
	}

	r.ServiceLevel.Delta.ErrorPercent = r.ServiceLevel.Actual.ErrorPercent - r.ServiceLevel.Target.ErrorPercent
	r.ServiceLevel.Delta.LatencyMs = r.ServiceLevel.Actual.LatencyMs - r.ServiceLevel.Target.LatencyMs

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
