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

func FromManifest(m *manifest.Manifest) (report *Report, err error) {
	// check for nil manifest
	if m == nil {
		err = go_errors.New("nil manifest.Manifest received")
		return
	}

	to := time.Now().UTC()

	var services []*Service = make([]*Service, 0)
	report = &Report{
		APIVersion: apiVersion,
		Timestamp:  timestampFn(),
		Services:   services,
	}

	for _, s := range m.Services {
		if s == nil {
			return nil, go_errors.New("I don't know anything about your services. Run `reliably slo init` to tell me.")
		}

		if len(s.ServiceLevels) == 0 {
			return nil, go_errors.New("you haven't told us about any SLOs, so we won't be able to give you a report. Sorry :(")
		}

		// Also need to check if there are SLIs in each Service Level
		sls := make([]*ServiceLevel, 0)
		rs := Service{
			Name:          s.Name,
			Dependencies:  []string{},
			ServiceLevels: sls,
		}

		for _, sl := range s.ServiceLevels {
			duration := sl.ObservationWindow.ToDuration()
			if duration == 0 {
				// we use a default value, in case we did not found any
				duration = oneDay
			} else {
				// We make sure to have duration rounded to 1-minute precision !!!
				// important for AWS metrics period being multiple of 60 sec !!!
				duration = duration.Truncate(time.Minute)
			}

			from := to.Add(-duration)

			allValues := []float64{}
			valuesHasError := false
			for _, sli := range sl.Indicators {
				provider, err := getProviderForResource(sli.Provider)
				if err != nil {
					return nil, fmt.Errorf("an error occurred while getting a provider for sli: %s => %s", sli.Provider, err)
				}
				defer provider.Close()

				var val float64
				switch sl.Type {
				case "latency":
					c := sl.Criteria.(manifest.LatencyCriteria)
					threshold := int(c.Threshold.Duration.Milliseconds())
					val, err = provider.GetLatencyAboveThresholdPercentage(sli.ID, from, to, threshold)
				case "availability":
					val, err = provider.GetAvailabilityPercentage(sli.ID, from, to)
				default:
					continue // skip unknown SL type - should not occur here though
				}

				if err == nil {
					allValues = append(allValues, val)
				} else {
					log.Debugf("an error occurred while getting %s data for resource: %s-%s => %v ", sl.Type, sli.Provider, sli.ID, err)
					valuesHasError = true
				}
			}

			var sloIsMet bool
			var objective float64 = sl.Objective
			var result *ServiceLevelResult
			if !valuesHasError {
				avg := average(allValues)
				delta := avg - objective
				sloIsMet = avg >= objective
				result = &ServiceLevelResult{
					Actual:   avg,
					Delta:    delta,
					SloIsMet: sloIsMet,
				}
			}
			rs.ServiceLevels = append(rs.ServiceLevels, &ServiceLevel{
				Name:      sl.Name,
				Type:      sl.Type,
				Objective: objective,
				Period:    sl.ObservationWindow,
				Result:    result,
				ObservationWindow: Window{
					To:   to,
					From: from,
				},
				errored: valuesHasError,
			})

		}

		if s.Dependencies != nil {
			rs.Dependencies = s.Dependencies
		}

		report.Services = append(report.Services, &rs)
	}

	return
}

func getProviderForResource(providerID metrics.ProviderType) (metrics.Provider, error) {
	if factory, ok := metrics.ProviderFactories[providerID]; ok {
		return factory()
	}

	return nil, fmt.Errorf("No provider factory found for '%s'", providerID)
}
