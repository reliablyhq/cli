package entities

var _ Entity = &ObjectiveResult{} // ensure the ObjectiveResult implements Entity interface
type ObjectiveResult struct {
	TypeMeta                `json:",inline" yaml:",inline"`
	ObjectiveResultResponse `json:",inline" yaml:",inline"`
}

type ObjectiveResultSpec struct {
	IndicatorSelector Selector `json:"indicatorSelector" yaml:"indicatorSelector"`
	ObjectivePercent  float64  `json:"objectivePercent" yaml:"objectivePercent"`
	ActualPercent     float64  `json:"actualPercent" yaml:"actualPercent"`
	RemainingPercent  float64  `json:"remainingPercent" yaml:"remainingPercent"`
}

func (o ObjectiveResult) Version() string {
	return o.TypeMeta.APIVersion
}

func (o ObjectiveResult) Kind() string {
	return o.TypeMeta.Kind
}

type ObjectiveResultResponse struct {
	Metadata `json:"metadata,omitempty"`
	Spec     ObjectiveResultSpec `json:"spec" yaml:"spec"`
}
