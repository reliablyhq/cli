package report

import (
	"reflect"
	"testing"
	"time"

	"github.com/reliablyhq/cli/core/metrics"
)

type dummyProvider struct{}

func (p *dummyProvider) Get99PercentLatencyMetricForResource(resourceID string, from, to time.Time) (float64, error) {
	return -1, nil
}

func (p *dummyProvider) GetErrorPercentageMetricForResource(resourceID string, from, to time.Time) (float64, error) {
	return -1, nil
}

func Test_getProviderForResource(t *testing.T) {
	p := &dummyProvider{}
	metrics.ProviderFactories["dummy"] = func() (metrics.Provider, error) { return p, nil }

	type args struct {
		ID string
	}
	tests := []struct {
		name    string
		args    args
		want    metrics.Provider
		wantErr bool
	}{
		{
			name:    "returns the correct provider",
			args:    args{ID: "dummy/123"},
			want:    p,
			wantErr: false,
		},
		{
			name:    "returns error if no provider was found",
			args:    args{ID: "xyz/123"},
			want:    nil,
			wantErr: true,
		},
		{
			name:    "returns error if no ID was supplied",
			args:    args{ID: ""},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getProviderForResource(tt.args.ID)
			if (err != nil) != tt.wantErr {
				t.Errorf("getProviderForResource() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getProviderForResource() = %v, want %v", got, tt.want)
			}
		})
	}
}
