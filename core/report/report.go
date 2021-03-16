package report

import "time"

type Report struct {
	Targets struct {
		ServiceLevel float32       `json:"service_level"`
		ErrorBudget  float32       `json:"error_budget"`
		Latency      time.Duration `json:"latency"`
	} `json:"target"`
	Threshold struct {
		Warning float32 `json:"warning"`
		Error   float32 `json:"error"`
	} `json:"threshold"`
	Delta struct {
		ServiceLevelPercent   float32 `json:"service_level_pc"`
		ErrorBudgetPercent    float32 `json:"error_budget_pc"`
		LatencyCeilingPercent float32 `json:"latency_ceiling_pc"`
	} `json:"delta"`
}
