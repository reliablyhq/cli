package metrics

import (
	"errors"
	"time"
)

type GcloudTrace struct {
	client struct{}
}

func NewGcloudTrace() (*GcloudTrace, error) {
	return nil, errors.New("NewAws not implemented")
}

func (p *GcloudTrace) GetLatencyMetricForResource(resourceID string, from, to time.Time) (float64, error) {
	return -1, errors.New("GetLatencyMetricForResource not implemented")
}

func (p *GcloudTrace) GetErrorPercentageMetricForResource(resourceID string, from, to time.Time) (float64, error) {
	return -1, errors.New("GetErrorPercentageMetricForResource not implemented")
}
