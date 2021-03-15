package advice

import (
	"fmt"

	"github.com/reliablyhq/cli/manifest"
)

func getSuggestionsForService(s *manifest.ServiceInfo) ([]Suggestion, error) {
	suggestions := make([]Suggestion, 0)

	if s.DesiredAvailability < 95 {
		suggestions = append(suggestions, "A desired availability of less than 95% should be possible for any well built app, deployed as a single instance")
	} else if s.DesiredAvailability < 99 {
		suggestions = append(suggestions, "A desired availability of less than 99% should be possible for any well built app, deployed with horizontal scaling in mind")
	} else {
		message := fmt.Sprintf("A desired availability of %v percent should be possible for any well built app, deployed with high availability in mind. That will involve load balancing the application, auto scaling it based on respource usage and controlling deployments to minimise downtime", s.DesiredAvailability)
		suggestions = append(suggestions, Suggestion(message))
	}

	return suggestions, nil
}
