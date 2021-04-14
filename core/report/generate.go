package report

import (
	go_errors "errors"
	"fmt"
	"time"

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
	// check for nil manifest
	if m == nil {
		return reports, fmt.Errorf("nil manifest.Manifest received")
	}

	for _, s := range m.Services {
		if s == nil {
			return nil, go_errors.New("I don't know anything about your services. Run `reliably slo init` to tell me.")
		}

		if len(s.ServiceLevels) == 0 {
			return nil, go_errors.New("you haven't told us about any SLOs, so we won't be able to give you a report. Sorry :(")
		}

		// Also need to check if there are SLIs in each Service Level

		r := Report{Name: s.Name}

		r.APIVersion = apiVersion
		r.Timestamp = timestampFn()
		r.Dependencies = []string{}

		allLatency := []float64{}
		allErrorPercentages := []float64{}

		var latencyHasErrors = false
		var errorPercentHasErrors = false

		for _, sl := range s.ServiceLevels {
			for _, sli := range sl.Indicators {
				provider, err := getProviderForResource(sli.Provider)
				if err != nil {
					return nil, fmt.Errorf("an error occured while getting a provider for sli: %s => %s", sli.Provider, err)
				}

				to := time.Now()
				from := to.Add(-oneDay)
				r.ObservationWindow.To = to
				r.ObservationWindow.From = from

				if l, err := provider.Get99PercentLatencyMetricForResource(sli.ID, from, to); err == nil {
					allLatency = append(allLatency, l)
				} else {
					log.Debugf("an error occured while getting latency data for resource: %s-%s => %v ", sli.Provider, sli.ID, err)
					latencyHasErrors = true
				}

				if e, err := provider.GetErrorPercentageMetricForResource(sli.ID, from, to); err == nil {
					allErrorPercentages = append(allErrorPercentages, e)
				} else {
	
					log.Debugf("an error occured while getting error percentage data for SLI: %s-%s => %v ", sli.Provider, sli.ID, err)
					errorPercentHasErrors = true
				}
			}

			// define actual indicator data received and errors
			actual := (&ServiceLevelIndicators{
				ErrorPercent: average(allErrorPercentages),
				LatencyMs:    int64(average(allLatency)),
			}).setErrorState(latencyErr, latencyHasErrors).
				setErrorState(errPercentErr, errorPercentHasErrors)

			r.ServiceLevel = &ServiceLevel{
				Target: &ServiceLevelIndicators{
					ErrorPercent: sl.Objective,
					LatencyMs:    sl.Threshold.Milliseconds(),
				},
				Actual: actual,
				Delta:  &ServiceLevelIndicators{},
			}

			r.ServiceLevel.Delta.ErrorPercent = r.ServiceLevel.Actual.ErrorPercent - r.ServiceLevel.Target.ErrorPercent
			r.ServiceLevel.Delta.LatencyMs = r.ServiceLevel.Actual.LatencyMs - r.ServiceLevel.Target.LatencyMs
		}

		if s.Dependencies != nil {
			r.Dependencies = s.Dependencies
		}

		reports = append(reports, &r)
	}

	return
}

func getProviderForResource(providerID string) (metrics.Provider, error) {
	if factory, ok := metrics.ProviderFactories[providerID]; ok {
		return factory()
	}

	return nil, fmt.Errorf("No provider factory found for '%s'", providerID)

}
