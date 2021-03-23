package report

import (
	"reflect"
	"testing"
	"time"

	"github.com/reliablyhq/cli/core/manifest"
	"github.com/reliablyhq/cli/core/metrics"
)

type dummyProvider struct {
	latencyMetricValue float64
	latencyError       error
	errorPercentValue  float64
	errorPercentError  error
}

func (p *dummyProvider) Get99PercentLatencyMetricForResource(resourceID string, from, to time.Time) (float64, error) {
	return p.latencyMetricValue, p.latencyError
}

func (p *dummyProvider) GetErrorPercentageMetricForResource(resourceID string, from, to time.Time) (float64, error) {
	return p.errorPercentValue, p.errorPercentError
}

func Test_getProviderForResource(t *testing.T) {
	p := &dummyProvider{}
	metrics.ProviderFactories["test_get_provider_for_resource"] = func() (metrics.Provider, error) { return p, nil }

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
			args:    args{ID: "test_get_provider_for_resource/123"},
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

func TestFromManifest(t *testing.T) {
	p := &dummyProvider{}
	metrics.ProviderFactories["test_from_manifest"] = func() (metrics.Provider, error) { return p, nil }

	type args struct {
		m *manifest.Manifest
	}
	tests := []struct {
		name    string
		args    args
		want    *Report
		wantErr bool
	}{
		{
			name:    "happy path",
			args:    args{},
			want:    &Report{},
			wantErr: false,
		},
		{
			name: "returns error if manifest is nil",
			args: args{
				m: nil,
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := FromManifest(tt.args.m)
			if (err != nil) != tt.wantErr {
				t.Errorf("FromManifest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FromManifest() = %v, want %v", got, tt.want)
			}
		})
	}
}
