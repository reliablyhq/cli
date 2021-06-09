package metrics

import (
	"time"

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

		Close() error // providers must explicitly be closed once not needed anymore
	}
)
