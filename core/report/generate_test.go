package report

import (
	"reflect"
	"testing"
	"time"

	"github.com/reliablyhq/cli/core"
	"github.com/reliablyhq/cli/core/manifest"
	"github.com/reliablyhq/cli/core/metrics"
	"github.com/stretchr/testify/assert"
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
		latencyMetricValue: 250,
		errorPercentValue:  99,
	}

	metrics.ProviderFactories["test_from_manifest"] = func() (metrics.Provider, error) { return p, nil }

	tVal := time.Now()
	timestampFn = func() time.Time { return tVal }

	type (
		args struct {
			m *manifest.Manifest
		}

		subtest struct {
			name    string
			args    args
			want    *Report
			wantErr bool
		}
	)

	tests := []subtest{
		{
			name: "returns report with correct info",
			args: args{
				m: &manifest.Manifest{
					Services: []*manifest.Service{
						&manifest.Service{
							Name: "Service A",
							ServiceLevels: []*manifest.ServiceLevel{
								&manifest.ServiceLevel{
									Name: "Service A Latency",
									Type: "latency",
									Threshold: core.Duration{Duration: 300 * time.Millisecond},
									Objective: 99,
									Indicators: []manifest.ServiceLevelIndicator{
										{
											Provider: "aws",
											ID: "arn2",
										},
									},
								},
							},
							Dependencies: []string{"dependencies"},
						},
					},
				},
			},
			want: &Report{
				Timestamp: tVal,
				ServiceLevel: &ServiceLevel{
					Target: &ServiceLevelIndicators{
						ErrorPercent: 99,
						LatencyMs:    300,
					},
					Actual: &ServiceLevelIndicators{
						ErrorPercent: p.errorPercentValue,
						LatencyMs:    int64(p.latencyMetricValue),
					},
					Delta: &ServiceLevelIndicators{
						ErrorPercent: p.errorPercentValue - 99,
						LatencyMs:    int64(p.latencyMetricValue) - 300,
					},
				},
				Dependencies: []string{"dependencies"},
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
			if (err != nil) && tt.wantErr {
				assert.Error(t, err, "FromManifest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// else just assert error
			assert.NoError(t, err)

			assert.Equal(t, got[0].Timestamp, tt.want.Timestamp,
				"FromManifest().Timestamp = %v, want %v", got[0].Timestamp, tt.want.Timestamp)

			assert.Equal(t, got[0].ServiceLevel, tt.want.ServiceLevel,
				"FromManifest().ServiceLevel = %v, want %v", got[0].ServiceLevel, tt.want.ServiceLevel)

			assert.Equal(t, got[0].Dependencies, tt.want.Dependencies,
				"FromManifest().Dependencies = %v, want %v", got[0].Dependencies, tt.want.Dependencies)

		})
	}
}
