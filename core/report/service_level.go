package report

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/reliablyhq/cli/manifest"
)

var (
	r = rand.New(rand.NewSource(time.Now().Unix()))
	d = 7 * 24 * time.Hour
)

const (
	thresholdPercent                  = 2
	noServiceLevel                    = "I don't know about the service level you are trying to provide. Run `reliably init service` to add a 'Service' block to your manifest."
	lessThan95pcAvailabilityMessage   = "An availability of less than 95% allows more than 36.5 hours of downtime per month, which should be possible for any well built app deployed as a single instance. This availability target is probably not high enough for a production-ready system."
	moreThan9999pcAvailabilityMessage = "An availabiliy of more than 99.99% is possible, but is quite expensive and may be unnecessary. 99.99% availability allows 4.38 minutes of downtime per month. You should make sure that less downtime is a requirement of your service."
	actualAvailabilityTooLowf         = "Current availability is lower than target availability by %.2f percent. Think about trying to increase the resources allocated of your application"
	actualAvailabilityTooHighf        = "Current availability is higher than target availability by %.2f percent. Think about reducing the resources allocated to your application - this could save you some money."
	errorBudgetExceededf              = "Error budget has been exceeded by %.2f percent. This is pretty bad :("
	errorBudgetTooLowf                = "You are considerably under your error budget by %.2f percent. You could tighten your budget, or could decrease the quality of the experience your application provides (e.g by reducing the amount of resources given to your application)."
)

func getSuggestionsForServiceLevel(m *manifest.Manifest) ([]Suggestion, error) {
	suggestions := make([]Suggestion, 0)

	if m.ServiceLevel == nil {
		suggestions = append(suggestions, noServiceLevel)
		return suggestions, nil
	}

	if m.ServiceLevel.Availability < 95 {
		suggestions = append(suggestions, lessThan95pcAvailabilityMessage)
	} else if m.ServiceLevel.Availability > 99.99 {
		suggestions = append(suggestions, moreThan9999pcAvailabilityMessage)
	}

	// todo: get current availability
	delta := getCurrentAvailability(m, d) - m.ServiceLevel.Availability
	if delta < 0 {
		s := fmt.Sprintf(actualAvailabilityTooLowf, delta)
		suggestions = append(suggestions, Suggestion(s))
	} else if delta > thresholdPercent {
		s := fmt.Sprintf(actualAvailabilityTooHighf, delta)
		suggestions = append(suggestions, Suggestion(s))
	}

	delta = getCurrentErrorPc(m, d) - m.ServiceLevel.ErrorBudgetPercent
	if delta < thresholdPercent {
		s := fmt.Sprintf(errorBudgetTooLowf, delta)
		suggestions = append(suggestions, Suggestion(s))
	} else if delta > thresholdPercent {
		s := fmt.Sprintf(errorBudgetExceededf, delta)
		suggestions = append(suggestions, Suggestion(s))
	}

	return suggestions, nil
}

func getCurrentAvailability(m *manifest.Manifest, d time.Duration) float32 {
	return 100 * r.Float32()
}

func getCurrentErrorPc(m *manifest.Manifest, d time.Duration) float32 {
	return 25 * r.Float32()
}
