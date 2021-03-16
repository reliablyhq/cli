package report

import "time"

type Report struct {
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
