package entities

import (
	"time"
)

var _ Entity = &Indicator{} // ensure the Indicator implements Entity interface

type Indicator struct {
	TypeMeta `json:",inline" yaml:",inline"`
	Metadata `json:"metadata,omitempty"`

	Spec IndicatorSpec `json:"spec" yaml:"spec"`
}

type IndicatorSpec struct {
	From    time.Time `json:"from"`
	To      time.Time `json:"to"`
	Percent float64   `json:"percent"`
}

func (i *Indicator) Version() string {
	return i.TypeMeta.APIVersion
}

func (i *Indicator) Kind() string {
	return i.TypeMeta.Kind
}

// NewIndicator returns a new instance of the Indicator struct with
// non-zero values for internal maps
func NewIndicator() *Indicator {
	return &Indicator{
		TypeMeta: TypeMeta{APIVersion: "reliably.com/v1", Kind: "Indicator"},
		Metadata: Metadata{
			Labels:    Labels{},
			RelatedTo: []map[string]string{},
		},
		Spec: IndicatorSpec{},
	}
}

func NewIndicatorForObjective(o *Objective, from time.Time, to time.Time) *Indicator {

	i := &Indicator{
		TypeMeta: TypeMeta{APIVersion: o.Version(), Kind: "Indicator"},
		Metadata: Metadata{
			Labels:    o.Spec.IndicatorSelector,
			RelatedTo: []map[string]string{},
		},
		Spec: IndicatorSpec{From: from, To: to},
	}

	// adding from/to to labels as well, to enforce unique-ness per indicator
	i.Metadata.Labels["from"] = i.Spec.From.String()
	i.Metadata.Labels["to"] = i.Spec.To.String()

	return i
}
