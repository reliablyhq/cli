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
		ServiceLevel       float32       `json:"service_level"`
		ErrorBudgetPercent float32       `json:"error_budget"`
		Latency            time.Duration `json:"latency"`
	} `json:"target"`

	Delta struct {
		ServiceLevel       float32       `json:"service_level"`
		ErrorBudgetPercent float32       `json:"error_budget_pc"`
		Latency            time.Duration `json:"latency"`
	} `json:"delta"`
}
