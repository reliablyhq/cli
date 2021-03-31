package metrics

import "time"

type DummyProvider struct{}

func (d *DummyProvider) Get99PercentLatencyMetricForResource(resourceID string, from, to time.Time) (float64, error) {
	return 500 * r.Float64(), nil
}

func (d *DummyProvider) GetErrorPercentageMetricForResource(resourceID string, from, to time.Time) (float64, error) {
	return 25 * r.Float64(), nil
}
