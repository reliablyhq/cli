package report

import (
	"time"

	"github.com/reliablyhq/cli/core/manifest"
)

func getCurrentAvailability(m *manifest.Manifest, d time.Duration) (float64, error) {
	return 100 * r.Float64(), nil
}

func getCurrentErrorPc(m *manifest.Manifest, d time.Duration) (float64, error) {
	return 25 * r.Float64(), nil
}

func get99PercentLatency(m *manifest.Manifest, d time.Duration) (time.Duration, error) {
	return 100 * time.Millisecond, nil
}
