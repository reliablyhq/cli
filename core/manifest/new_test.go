package manifest

import (
	"reflect"
	"testing"
)

func TestEmpty(t *testing.T) {
	tests := []struct {
		name string
		want *Manifest
	}{
		{
			name: "happy path",
			want: &Manifest{
				ServiceLevel: &ServiceLevel{},
				CI:           &ContinuousIntegrationInfo{},
				Hosting:      &Hosting{},
				IAC:          &IAC{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Empty(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Empty() = %v, want %v", got, tt.want)
			}
		})
	}
}
