package metrics

import (
	"math/rand"
	"time"
)

var r = rand.New(rand.NewSource(time.Now().Unix()))

var ProviderFactories = map[string]ProviderFactory{
	"dummy": func() (Provider, error) { return &DummyProvider{}, nil },
}

type (
	ProviderFactory func() (Provider, error)

	Provider interface {
		Get99PercentLatencyMetricForResource(resourceID string, from, to time.Time) (float64, error)
		GetErrorPercentageMetricForResource(resourceID string, from, to time.Time) (float64, error)
	}
)

type DummyProvider struct{}

func (d *DummyProvider) Get99PercentLatencyMetricForResource(resourceID string, from, to time.Time) (float64, error) {
	return 500 * r.Float64(), nil
}

func (d *DummyProvider) GetErrorPercentageMetricForResource(resourceID string, from, to time.Time) (float64, error) {
	return 25 * r.Float64(), nil
}
