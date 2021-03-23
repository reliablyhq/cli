package report

import (
	"time"

	"github.com/reliablyhq/cli/core/manifest"
)

func getCurrentAvailability(m *manifest.Manifest, _ time.Duration) float64 {
	return 100 * r.Float64()
}

func getCurrentErrorPc(m *manifest.Manifest, _ time.Duration) float64 {
	return 20 * r.Float64()
}

func get99PercentLatencyMs(m *manifest.Manifest, _ time.Duration) time.Duration {
	return time.Duration(r.Int31n(2000)) * time.Millisecond
}
