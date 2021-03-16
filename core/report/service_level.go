package report

import (
	"github.com/reliablyhq/cli/manifest"
)

const (
	lessThan95pcAvailabilityMessage   = "An availability of less than 95% allows more than 36.5 hours of downtime per month, which should be possible for any well built app deployed as a single instance"
	moreThan9999pcAvailabilityMessage = "An availabiliy of more than 99.99% is possible, but is quite expensive and may be unnecessary. 99.99% availability allows 4.38 minutes of downtime per month. You should make sure that less downtime is a requirement of your service."
)

func getSuggestionsForServiceLevel(m *manifest.Manifest) ([]Suggestion, error) {
	suggestions := make([]Suggestion, 0)

	if m.ServiceLevel.Availability < 95 {
		suggestions = append(suggestions, lessThan95pcAvailabilityMessage)
	} else if m.ServiceLevel.Availability > 99.99 {
		suggestions = append(suggestions, moreThan9999pcAvailabilityMessage)
	}

	// todo: get current availability

	return suggestions, nil
}
