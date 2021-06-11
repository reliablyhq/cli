package metrics

import (
	"time"

	"github.com/reliablyhq/cli/core/entities"
	"github.com/reliablyhq/cli/core/metrics/aws"
	"github.com/reliablyhq/cli/core/metrics/gcp"
)

const (
	AWSProvider ProviderType = "aws"
	GCPProvider ProviderType = "gcp"
)

var ProviderFactories = map[ProviderType]ProviderFactory{
	"aws": func() (Provider, error) { return aws.NewAwsCloudWatch() },
	"gcp": func() (Provider, error) { return gcp.NewGCP() },
}

type (
	ProviderType    string
	ProviderFactory func() (Provider, error)

	Provider interface {
		Get99PercentLatencyMetricForResource(resourceID string, from, to time.Time) (float64, error)
		GetErrorPercentageMetricForResource(resourceID string, from, to time.Time) (float64, error)

		GetLatencyAboveThresholdPercentage(resourceID string, from, to time.Time, threshold int) (float64, error)
		GetAvailabilityPercentage(resourceID string, from, to time.Time) (float64, error)

		// Note that this function must return an empty
		// string if resource cannot be determined
		ResourceFromSelector(entities.Selector) string
		Close() error // providers must explicitly be closed once not needed anymore

	}
)

// FromSelector - identifies provider from Objective
// returns nil if provider cannot be identified
func FromSelector(obj *entities.Selector) Provider {
	return nil
}

// ResourceFromSelector - identifies the resource ID from the
// given selector
// returns empty string if resource cannot be determined
func ResourceIDFromSelector(obj *entities.Selector) string {
	return ""
}
