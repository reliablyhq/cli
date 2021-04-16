package report

import (
	//"fmt"
	"time"
)

type Report struct {
	APIVersion string     `json:"api_version"`
	Timestamp  time.Time  `json:"timestamp"`
	Services   []*Service `json:"services"`
}

type Service struct {
	Dependencies  []string        `json:"-"`
	ServiceLevels []*ServiceLevel `json:"service_levels"`
	Name          string          `json:"name"`
}

type ServiceLevel struct {
	Name      string              `json:"name"`
	Type      string              `json:"type"`
	Objective float64             `json:"objective"`
	Result    *ServiceLevelResult `json:"result"`

	//Target *ServiceLevelIndicators `json:"target"`
	//Actual *ServiceLevelIndicators `json:"actual"`
	//Delta  *ServiceLevelIndicators `json:"delta"`

	// TODO: decide whether this should be implemented as
	// Observation Windor or Boundary
	ObservationWindow Window `json:"window"`

	// used to record whether a particular SLI had
	// error in it's retrieval process
	errored bool `json:"-"`

	// used to record whether the SLO is met or not
	sloIsMet bool `json:"-"`
}

type ServiceLevelIndicators struct {
	ErrorPercent   float64 `json:"error_percent"`
	LatencyPercent int64   `json:"latency_percent"`

	// used to record whether a particular property had
	// error in it's retrieval process
	errored bool `json:"-"`
}

type Window struct {
	From time.Time `json:"from"`
	To   time.Time `json:"to"`
}

type ServiceLevelResult struct {
	//Objective interface{} `json:"slo"`
	Actual interface{} `json:"actual"`
	Delta  interface{} `json:"delta"`
}

/*
func (s *ServiceLevelIndicators) setErrorState(i indicatorErrType, b bool) *ServiceLevelIndicators {
	s.errored[i] = b
	return s
}

func (s *ServiceLevelIndicators) hasErrors(i indicatorErrType) bool {
	return s.errored[i]
}

func (s *ServiceLevelIndicators) errorPercentString() string {
	if s.hasErrors(errPercentErr) {
		return "---"
	}
	return fmt.Sprintf("%.2f", s.ErrorPercent)
}

func (s *ServiceLevelIndicators) latencyMsString() string {
	if s.hasErrors(latencyErr) {
		return "---"
	}
	return fmt.Sprintf("%dms", s.LatencyMs)
}
*/
// indicatorErrType - used to define indicator error types
type indicatorErrType int

var latencyErr indicatorErrType = 0
var errPercentErr indicatorErrType = 1
