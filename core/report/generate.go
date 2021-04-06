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
	oneHour    = 1 * time.Hour
	oneDay     = 24 * oneHour
	oneWeek    = 7 * oneDay
	apiVersion = "1.0rc"
)

var (
	timestampFn func() time.Time = time.Now().UTC
)

var notSuportedErrorBuilder = func(thingType, thingName string) error {
	return fmt.Errorf("%s type '%s' is not currently supported. I've informed ReliablyHQ about this; check back later - maybe we'll be able to help then.", thingType, thingName)
}

func FromManifest(m *manifest.Manifest) (reports []*Report, err error) {
	for _, s := range m.ServiceLevel {
		if s == nil {
			return nil, go_errors.New("I don't know anything about your service objectives. Run `reliably slo init` to tell me.")
		}

		if len(s.Resources) == 0 {
			return nil, go_errors.New("you haven't told us about any resources, so we won't be able to give you a report. Sorry :(")
		}

		r := Report{Name: s.Name}
		allErrors := make([]error, 0)

		r.APIVersion = apiVersion
		r.Timestamp = timestampFn()
		r.Dependencies = []string{}

		allLatency := []float64{}
		allErrorPercentages := []float64{}

		for _, resource := range s.Resources {
			provider, err := getProviderForResource(resource.Provider)
			if err != nil {
				log.Errorf("an error occured while getting a provider for resource: %s => %s", resource.Provider, err)
				continue
			}

			to := time.Now()
			from := to.Add(-oneDay)
			r.ObservationWindow.To = to
			r.ObservationWindow.From = from

			if l, err := provider.Get99PercentLatencyMetricForResource(resource.ID, from, to); err == nil {
				allLatency = append(allLatency, l)
			} else {
				log.Errorf("an error occured while getting latency data for resource: %s-%s => %v ", resource.Provider, resource.ID, err)
			}

			if e, err := provider.GetErrorPercentageMetricForResource(resource.ID, from, to); err == nil {
				allErrorPercentages = append(allErrorPercentages, e)
			} else {
				log.Errorf("an error occured while getting error percentage data for resource: %s-%s => %v ", resource.Provider, resource.ID, err)
			}

		}

		r.ServiceLevel = &ServiceLevel{
			Target: &ServiceLevelIndicators{
				ErrorPercent: s.Objective.ErrorBudgetPercent,
				LatencyMs:    s.Objective.Latency.Milliseconds(),
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
			return reports, errors.NewCompoundError("multiple errors occured", allErrors)
		}

		reports = append(reports, &r)
	}

	return
}

// func FromManifest(m *manifest.Manifest) (*Report, error) {
// 	if m == nil {
// 		return nil, go_errors.New("manifest is nil")
// 	}

// 	r := Report{}
// 	allErrors := make([]error, 0)

// 	r.APIVersion = apiVersion
// 	r.Timestamp = timestampFn()
// 	r.Dependencies = []string{}

// 	allLatency := []float64{}
// 	allErrorPercentages := []float64{}

// 	if m.ServiceLevel == nil {
// 		return nil, go_errors.New("I don't know anything about your service objectives. Run `reliably slo init` to tell me.")
// 	}

// 	if len(m.ServiceLevel.Resources) == 0 {
// 		return nil, go_errors.New("you haven't told us about any resources, so we won't be able to give you a report. Sorry :(")
// 	}

// 	for _, resource := range m.ServiceLevel.Resources {
// 		provider, err := getProviderForResource(resource.Provider)
// 		if err != nil {
// 			log.Errorf("an error occured while getting a provider for resource: %s => %s", resource.Provider, err)
// 			continue
// 		}

// 		to := time.Now()
// 		from := to.Add(-oneDay)
// 		r.ObservationWindow.To = to
// 		r.ObservationWindow.From = from

// 		if l, err := provider.Get99PercentLatencyMetricForResource(resource.ID, from, to); err == nil {
// 			allLatency = append(allLatency, l)
// 		} else {
// 			log.Errorf("an error occured while getting latency data for resource: %s-%s => %v ", resource.Provider, resource.ID, err)
// 		}

// 		if e, err := provider.GetErrorPercentageMetricForResource(resource.ID, from, to); err == nil {
// 			allErrorPercentages = append(allErrorPercentages, e)
// 		} else {
// 			log.Errorf("an error occured while getting error percentage data for resource: %s-%s => %v ", resource.Provider, resource.ID, err)
// 		}

// 	}

// 	r.ServiceLevel = &ServiceLevel{
// 		Target: &ServiceLevelIndicators{
// 			ErrorPercent: m.ServiceLevel.Objective.ErrorBudgetPercent,
// 			LatencyMs:    m.ServiceLevel.Objective.Latency.Milliseconds(),
// 		},
// 		Actual: &ServiceLevelIndicators{
// 			ErrorPercent: average(allErrorPercentages),
// 			LatencyMs:    int64(average(allLatency)),
// 		},
// 		Delta: &ServiceLevelIndicators{},
// 	}

// 	r.ServiceLevel.Delta.ErrorPercent = r.ServiceLevel.Actual.ErrorPercent - r.ServiceLevel.Target.ErrorPercent
// 	r.ServiceLevel.Delta.LatencyMs = r.ServiceLevel.Actual.LatencyMs - r.ServiceLevel.Target.LatencyMs

// 	if m.Dependencies != nil {
// 		r.Dependencies = m.Dependencies
// 	}

// 	if len(allErrors) > 0 {
// 		return &r, errors.NewCompoundError("multiple errors occured", allErrors)
// 	}

// 	return &r, nil
// }

func getProviderForResource(providerID string) (metrics.Provider, error) {
	if factory, ok := metrics.ProviderFactories[providerID]; ok {
		return factory()
	}

	return nil, fmt.Errorf("No provider factory found for '%s'", providerID)

}
