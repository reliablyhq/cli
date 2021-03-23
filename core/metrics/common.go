package metrics

import (
	"time"
)

type (
	Provider interface {
		GetAverageLatencyMetricForResource(resourceID string, from, to time.Time) (float64, error)
		GetErrorPercentageMetricForResource(resourceID string, from, to time.Time) (float64, error)
	}
)
