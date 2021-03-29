package report

import (
	"time"
)

type Report struct {
	APIVersion      string        `json:"api_version"`
	ApplicationName string        `json:"application_name"`
	Timestamp       time.Time     `json:"timestamp"`
	Dependencies    []string      `json:"dependencies"`
	ServiceLevel    *ServiceLevel `json:"service_level"`
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
