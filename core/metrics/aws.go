package metrics

import (
	"errors"
	"time"
)

type Aws struct{}

func NewAws() (*Aws, error) {
	return nil, errors.New("NewAws not implemented")
}

func (p *Aws) GetLatencyMetricForResource(resourceID string, from, to time.Time) (float64, error) {
	return -1, errors.New("GetLatencyMetricForResource not implemented")
}

func (p *Aws) GetErrorPercentageMetricForResource(resourceID string, from, to time.Time) (float64, error) {
	return -1, errors.New("GetErrorPercentageMetricForResource not implemented")
}
