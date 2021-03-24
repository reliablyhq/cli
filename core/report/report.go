package report

import (
	"time"

	"github.com/reliablyhq/cli/core/manifest"
)

type Report struct {
	ApplicationName string              `json:"application_name"`
	Timestamp       time.Time           `json:"timestamp"`
	Dependencies    []*manifest.AppInfo `json:"dependencies"`
	ServiceLevel    *ServiceLevel       `json:"service_level"`
}

type ServiceLevel struct {
	Target *ServiceLevelIndicators `json:"target"`
	Actual *ServiceLevelIndicators `json:"actual"`
	Delta  *ServiceLevelIndicators `json:"delta"`
}

type ServiceLevelIndicators struct {
	ErrorBudgetPercent float64 `json:"error_budget"`
	LatencyMs          int64   `json:"latency_ms"`
}
