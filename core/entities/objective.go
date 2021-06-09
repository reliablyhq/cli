package entities

import (
	"github.com/reliablyhq/cli/core"
)

var _ Entity = &Objective{} // ensure the Objective implements Entity interface
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

func (o Objective) Version() string {
	return o.TypeMeta.APIVersion
}

func (o Objective) Kind() string {
	return o.TypeMeta.Kind
}

// NewObjective returns a new instance of the Objective struct with
// non-zero values for internal maps
func NewObjective() *Objective {
	return &Objective{
		TypeMeta: TypeMeta{APIVersion: "v1", Kind: "Objective"},
		Metadata: Metadata{
			Labels:    Labels{},
			RelatedTo: []map[string]string{},
		},
		Spec: ObjectiveSpec{
			IndicatorSelector: Selector{},
		},
	}
}
