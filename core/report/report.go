package report

import (
	"time"

	"github.com/reliablyhq/cli/core/manifest"
)

type Report struct {
	ApplicationName string                 `json:"application_name"`
	Timestamp       time.Time              `json:"timestamp"`
	Dependencies    []*manifest.Dependency `json:"dependencies"`

	Targets struct {
		ServiceLevel       float64 `json:"service_level"`
		ErrorBudgetPercent float64 `json:"error_budget"`
		LatencyMs          int64   `json:"latency_ms"`
	} `json:"target"`

	Delta struct {
		ServiceLevel       float64 `json:"service_level"`
		ErrorBudgetPercent float64 `json:"error_budget_pc"`
		LatencyMs          int64   `json:"latency_ms"`
	} `json:"delta"`
}
