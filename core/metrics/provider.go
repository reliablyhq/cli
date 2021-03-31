package metrics

import (
	"math/rand"
	"time"

	"github.com/reliablyhq/cli/core/metrics/aws"
	"github.com/reliablyhq/cli/core/metrics/gcp"
	"github.com/reliablyhq/cli/version"
)

var r = rand.New(rand.NewSource(time.Now().Unix()))

var ProviderFactories = map[string]ProviderFactory{
	"aws": func() (Provider, error) { return aws.NewAwsCloudWatch() },
	"gcp": func() (Provider, error) { return gcp.NewGCP() },
}

type (
	ProviderFactory func() (Provider, error)

	Provider interface {
		Get99PercentLatencyMetricForResource(resourceID string, from, to time.Time) (float64, error)
		GetErrorPercentageMetricForResource(resourceID string, from, to time.Time) (float64, error)
	}
)

func init() {
	if version.IsDevVersion() {
		ProviderFactories["dummy"] = func() (Provider, error) { return &DummyProvider{}, nil }
	}
}
