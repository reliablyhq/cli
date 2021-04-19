package metrics

import (
	"time"

	"github.com/reliablyhq/cli/core/metrics/aws"
	"github.com/reliablyhq/cli/core/metrics/gcp"
)

var ProviderFactories = map[string]ProviderFactory{
	"aws": func() (Provider, error) { return aws.NewAwsCloudWatch() },
	"gcp": func() (Provider, error) { return gcp.NewGCP() },
}

type (
	ProviderFactory func() (Provider, error)

	Provider interface {
		Get99PercentLatencyMetricForResource(resourceID string, from, to time.Time) (float64, error)
		GetErrorPercentageMetricForResource(resourceID string, from, to time.Time) (float64, error)
		GetLatencyAboveThresholdPercentage(resourceID string, threshold int, from, to time.Time) (float64, error)
	}
)
