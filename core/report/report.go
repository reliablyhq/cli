package report

import (
	//"fmt"
	"time"

	"github.com/reliablyhq/cli/core"
)

type Report struct {
	APIVersion string     `json:"-" yaml:"-"`
	Timestamp  time.Time  `json:"timestamp" yaml:"timestamp"`
	Services   []*Service `json:"services" yaml:"services"`
}

type Service struct {
	Dependencies  []string        `json:"-" yaml:"-"`
	ServiceLevels []*ServiceLevel `json:"service_levels" yaml:"service_levels"`
	Name          string          `json:"name" yaml:"name"`
}

type ServiceLevel struct {
	Name      string               `json:"name" yaml:"name"`
	Type      string               `json:"type" yaml:"type"`
	Objective float64              `json:"objective" yaml:"objective"`
	Period    core.Iso8601Duration `yaml:"period,omitempty" json:"period,omitempty"`
	Result    *ServiceLevelResult  `json:"result" yaml:"result"`

	//Target *ServiceLevelIndicators `json:"target"`
	//Actual *ServiceLevelIndicators `json:"actual"`
	//Delta  *ServiceLevelIndicators `json:"delta"`

	// TODO: decide whether this should be implemented as
	// Observation Windor or Boundary
	ObservationWindow Window `json:"window" yaml:"window"`

	// used to record whether a particular SLI had
	// error in it's retrieval process
	errored bool `json:"-" yaml:"-"`
}

type Window struct {
	From time.Time `json:"from"`
	To   time.Time `json:"to"`
}

type ServiceLevelResult struct {
	Actual interface{} `json:"actual"`
	Delta  interface{} `json:"delta"`

	// used to record whether the SLO is met or not
	SloIsMet bool `json:"slo_is_met"`
}

// GetResult returns the result for a given service & slo name
func (r *Report) GetResult(svcName string, sloName string) *ServiceLevelResult {

	for _, svc := range r.Services {
		if svc.Name != svcName {
			continue
		}

		for _, slo := range svc.ServiceLevels {
			if slo.Name != sloName {
				continue
			}

			return slo.Result
		}
	}

	return nil
}
