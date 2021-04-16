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

	to := time.Now()
	from := to.Add(-oneDay)

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
		//r := Report{Name: s.Name, ServiceLevels: sls}
		rs := Service{
			Name:          s.Name,
			Dependencies:  []string{},
			ServiceLevels: sls,
		}

		//allLatency := []float64{}
		//allErrorPercentages := []float64{}

		allValues := []float64{}
		valuesHasError := false

		//var latencyHasErrors = false
		//var errorPercentHasErrors = false

		for _, sl := range s.ServiceLevels {

			//fmt.Println(sl.Name, sl.Type, sl.Objective, sl.Threshold)

			for _, sli := range sl.Indicators {
				provider, err := getProviderForResource(sli.Provider)
				if err != nil {
					return nil, fmt.Errorf("an error occured while getting a provider for sli: %s => %s", sli.Provider, err)
				}

				if sl.Type == "latency" {
					if l, err := provider.Get99PercentLatencyMetricForResource(sli.ID, from, to); err == nil {
						//allLatency = append(allLatency, l)
						allValues = append(allValues, l)
					} else {
						log.Debugf("an error occured while getting latency data for resource: %s-%s => %v ", sli.Provider, sli.ID, err)
						//latencyHasErrors = true
						valuesHasError = true
					}
				}

				if sl.Type == "availability" {
					if e, err := provider.GetErrorPercentageMetricForResource(sli.ID, from, to); err == nil {
						//allErrorPercentages = append(allErrorPercentages, e)
						allValues = append(allValues, e)
					} else {

						log.Debugf("an error occured while getting error percentage data for SLI: %s-%s => %v ", sli.Provider, sli.ID, err)
						//errorPercentHasErrors = true
						valuesHasError = true
					}
				}

			}

			/*
				// define actual indicator data received and errors
				actual := (&ServiceLevelIndicators{
					ErrorPercent: average(allErrorPercentages),
					LatencyMs:    int64(average(allLatency)),
				}).setErrorState(latencyErr, latencyHasErrors).
					setErrorState(errPercentErr, errorPercentHasErrors)
			*/

			//fmt.Println(">>>", sl.Name, sl.Type, allValues, valuesHasError)

			var sloIsMet bool
			var objective float64 = sl.Objective
			var result *ServiceLevelResult
			if !valuesHasError {

				// hack for now, until we merge the new API for fetching latency as %

				if sl.Type == "latency" {
					objective = float64(sl.Threshold.Duration.Milliseconds())
					avg := average(allValues)
					delta := avg - objective
					sloIsMet = avg <= objective

					result = &ServiceLevelResult{
						//Objective: objective,
						Actual:   round2digits(avg),
						Delta:    round2digits(delta),
						sloIsMet: sloIsMet,
					}

				} else {
					avg := average(allValues)
					delta := avg - objective
					sloIsMet = avg >= objective

					result = &ServiceLevelResult{
						//Objective: sl.Objective,
						Actual:   avg,
						Delta:    delta,
						sloIsMet: sloIsMet,
					}
				}

			}
			rs.ServiceLevels = append(rs.ServiceLevels, &ServiceLevel{
				Name:      sl.Name,
				Type:      sl.Type,
				Objective: objective,
				Result:    result,
				ObservationWindow: Window{
					To:   to,
					From: from,
				},
				errored: valuesHasError,
			})

			//r.ServiceLevel.Delta.ErrorPercent = r.ServiceLevel.Actual.ErrorPercent - r.ServiceLevel.Target.ErrorPercent
			//r.ServiceLevel.Delta.LatencyMs = r.ServiceLevel.Actual.LatencyMs - r.ServiceLevel.Target.LatencyMs
		}

		if s.Dependencies != nil {
			rs.Dependencies = s.Dependencies
		}

		//reports = append(reports, &r)
		report.Services = append(report.Services, &rs)
	}

	return
}

func getProviderForResource(providerID string) (metrics.Provider, error) {
	if factory, ok := metrics.ProviderFactories[providerID]; ok {
		return factory()
	}

	return nil, fmt.Errorf("No provider factory found for '%s'", providerID)

}
