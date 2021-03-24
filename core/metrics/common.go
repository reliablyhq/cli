package metrics

import (
	"time"
)

var ProviderFactories = map[string]ProviderFactory{
	"gcp": func() (Provider, error) { return NewGCP() },
}

type (
	ProviderFactory func() (Provider, error)

	Provider interface {
		Get99PercentLatencyMetricForResource(resourceID string, from, to time.Time) (float64, error)
		GetErrorPercentageMetricForResource(resourceID string, from, to time.Time) (float64, error)
	}
)
