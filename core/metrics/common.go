package metrics

import (
	"time"
)

type (
	Provider interface {
		Get99PercentLatencyMetricForResource(resourceID string, from, to time.Time) (float64, error)
		GetErrorPercentageMetricForResource(resourceID string, from, to time.Time) (float64, error)
	}
)
