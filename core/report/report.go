package report

import (
	"fmt"
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

	// used to record whether a particular property had
	// error in it's retrieval process
	errored [2]bool `json:"-"`
}

func (s *ServiceLevelIndicators) setErrorState(i indicatorErrType, b bool) *ServiceLevelIndicators {
	s.errored[i] = b
	return s
}

func (s *ServiceLevelIndicators) hasErrors(i indicatorErrType) bool {
	return s.errored[i]
}

func (s *ServiceLevelIndicators) errorPerentString() string {
	if s.errored[errPercentErr] {
		return fmt.Sprintf("%.2f", s.ErrorPercent)
	}
	return "---"
}

func (s *ServiceLevelIndicators) latencyMsString() string {
	if s.errored[latencyErr] {
		return fmt.Sprintf("%dms", s.LatencyMs)
	}
	return "---"
}

// indicatorErrType - used to define indicator error types
type indicatorErrType int

var latencyErr indicatorErrType = 0
var errPercentErr indicatorErrType = 1
