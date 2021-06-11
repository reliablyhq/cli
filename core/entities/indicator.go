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
