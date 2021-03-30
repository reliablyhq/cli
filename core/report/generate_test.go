package report

import (
	"reflect"
	"testing"
	"time"

	"github.com/reliablyhq/cli/core"
	"github.com/reliablyhq/cli/core/manifest"
	"github.com/reliablyhq/cli/core/metrics"
)

type dummyProvider struct {
	latencyMetricValue float64
	latencyError       error
	errorPercentValue  float64
	errorPercentError  error
	resourceID         string
}

func (p *dummyProvider) Get99PercentLatencyMetricForResource(resourceID string, from, to time.Time) (float64, error) {
	p.resourceID = resourceID
	return p.latencyMetricValue, p.latencyError
}

func (p *dummyProvider) GetErrorPercentageMetricForResource(resourceID string, from, to time.Time) (float64, error) {
	p.resourceID = resourceID
	return p.errorPercentValue, p.errorPercentError
}

func Test_getProviderForResource(t *testing.T) {
	p := &dummyProvider{}
	metrics.ProviderFactories["test_get_provider_for_resource"] = func() (metrics.Provider, error) { return p, nil }

	type args struct {
		providerID string
	}
	tests := []struct {
		name    string
		args    args
		want    metrics.Provider
		wantErr bool
	}{
		{
			name:    "returns the correct provider",
			args:    args{providerID: "test_get_provider_for_resource"},
			want:    p,
			wantErr: false,
		},
		{
			name:    "returns error if no provider was found",
			args:    args{providerID: "xyz"},
			want:    nil,
			wantErr: true,
		},
		{
			name:    "returns error if no ID was supplied",
			args:    args{providerID: ""},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getProviderForResource(tt.args.providerID)
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
	p := &dummyProvider{
		latencyMetricValue: 100,
		errorPercentValue:  1,
	}
	metrics.ProviderFactories["test_from_manifest"] = func() (metrics.Provider, error) { return p, nil }

	tVal := time.Now()
	timestampFn = func() time.Time { return tVal }

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
			name: "returns report with correct info",
			args: args{
				m: &manifest.Manifest{
					App: &manifest.AppInfo{
						Name:       "test app",
						Owner:      "test owner",
						Repository: "test repo",
					},
					Service: &manifest.Service{
						Objective: &manifest.ServiceLevelObjective{
							Latency:            core.Duration{Duration: time.Millisecond * 250},
							ErrorBudgetPercent: 2.5,
						},
						Resources: []*manifest.ServiceResource{
							{
								Provider: "test_from_manifest",
								ID:       "abc13",
							},
						},
					},
					Dependencies: []string{"abc"},
				},
			},
			want: &Report{
				ApplicationName: "test app",
				Timestamp:       tVal,
				ServiceLevel: &ServiceLevel{
					Target: &ServiceLevelIndicators{
						ErrorPercent: 2.5,
						LatencyMs:    250,
					},
					Actual: &ServiceLevelIndicators{
						ErrorPercent: p.errorPercentValue,
						LatencyMs:    int64(p.latencyMetricValue),
					},
					Delta: &ServiceLevelIndicators{
						ErrorPercent: p.errorPercentValue - 2.5,
						LatencyMs:    int64(p.latencyMetricValue) - 250,
					},
				},
				Dependencies: []string{"abc"},
			},
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
			} else if err != nil && tt.wantErr {
				return
			}

			if tt.want.ApplicationName != got.ApplicationName {
				t.Errorf("wanted ApplicationName to be %s but was %s", tt.want.ApplicationName, got.ApplicationName)
				return
			}

			if !reflect.DeepEqual(got.Timestamp, tt.want.Timestamp) {
				t.Errorf("FromManifest().Timestamp = %v, want %v", got.Timestamp, tt.want.Timestamp)
				return
			}

			if !reflect.DeepEqual(got.ServiceLevel, tt.want.ServiceLevel) {
				t.Errorf("FromManifest().ServiceLevel = %v, want %v", got.ServiceLevel, tt.want.ServiceLevel)
				return
			}

			if !reflect.DeepEqual(got.Dependencies, tt.want.Dependencies) {
				t.Errorf("FromManifest().Dependencies = %v, want %v", got.Dependencies, tt.want.Dependencies)
				return
			}
		})
	}
}
