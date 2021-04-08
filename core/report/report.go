package report

import (
	"time"
)

type Report struct {
	APIVersion   string        `json:"api_version"`
	Timestamp    time.Time     `json:"timestamp"`
	Dependencies []string      `json:"dependencies"`
	ServiceLevel *ServiceLevel `json:"service_level"`
	Name         string        `json:"name"`

	// TODO: decide whether this should be implemented as
	// Observation Windor or Boundary
	ObservationWindow Window `json:"window"`
}

type Window struct {
	From time.Time `json:"from"`
	To   time.Time `json:"to"`
}

type ServiceLevel struct {
	Target *ServiceLevelIndicators `json:"target"`
	Actual *ServiceLevelIndicators `json:"actual"`
	Delta  *ServiceLevelIndicators `json:"delta"`
}

type ServiceLevelIndicators struct {
	ErrorPercent float64 `json:"error_percent"`
	LatencyMs    int64   `json:"latency_ms"`
}
