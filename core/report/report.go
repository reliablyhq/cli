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
}

type Window struct {
	From time.Time `json:"from"`
	To   time.Time `json:"to"`
}

type ServiceLevelResult struct {
	Actual interface{} `json:"actual"`
	Delta  interface{} `json:"delta"`

	// used to record whether the SLO is met or not
	sloIsMet bool `json:"-"`
}
