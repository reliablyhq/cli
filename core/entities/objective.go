package entities

import (
	"github.com/reliablyhq/cli/core"
)

type Objective struct {
	TypeMeta `json:",inline" yaml:",inline"`
	Metadata `json:"metadata,omitempty"`

	Spec ObjectiveSpec `json:"spec" yaml:"spec"`
}

type ObjectiveSpec struct {
	IndicatorSelector Selector      `json:"indicatorSelector" yaml:"indicatorSelector"`
	ObjectivePercent  float64       `json:"objectivePercent" yaml:"objectivePercent"`
	Window            core.Duration `json:"window" yaml:"window"`
}
