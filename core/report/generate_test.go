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
	latencyMetricValue  float64
	latencyError        error
	errorPercentValue   float64
	errorPercentError   error
	latencyPercentValue float64
	latencyPercentError error
	resourceID          string
}

func (p *dummyProvider) Close() error {
	return nil
}

func (p *dummyProvider) Get99PercentLatencyMetricForResource(resourceID string, from, to time.Time) (float64, error) {
	p.resourceID = resourceID
	return p.latencyMetricValue, p.latencyError
}

func (p *dummyProvider) GetErrorPercentageMetricForResource(resourceID string, from, to time.Time) (float64, error) {
	p.resourceID = resourceID
	return p.errorPercentValue, p.errorPercentError
}

func (p *dummyProvider) GetLatencyAboveThresholdPercentage(resourceID string, from, to time.Time, threshold int) (float64, error) {
	p.resourceID = resourceID
	return p.latencyPercentValue, p.latencyPercentError
}

func (p *dummyProvider) GetAvailabilityPercentage(resourceID string, from, to time.Time) (float64, error) {
	if p.errorPercentError != nil {
		return -1, p.errorPercentError
	}
	return 100.0 - p.errorPercentValue, nil
}

func Test_getProviderForResource(t *testing.T) {
	p := &dummyProvider{}
	metrics.ProviderFactories["test_get_provider_for_resource"] = func() (metrics.Provider, error) { return p, nil }

	type args struct {
		providerID metrics.ProviderType
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
		latencyPercentValue: 93,
		errorPercentValue:   99,
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
						{
							Name: "Service A",
							ServiceLevels: []*manifest.ServiceLevel{
								{
									Name: "Service A Latency",
									Type: "latency",
									Criteria: manifest.LatencyCriteria{
										Threshold: core.Duration{Duration: 300 * time.Millisecond},
									},
									Objective: 99,
									Indicators: []manifest.ServiceLevelIndicator{
										{
											Provider: "test_from_manifest",
											ID:       "abc13",
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
				Services: []*Service{
					{
						Name:         "Service A",
						Dependencies: []string{"dependencies"},
						ServiceLevels: []*ServiceLevel{
							{
								Name:      "Service A Latency",
								Type:      "latency",
								Objective: 99,
								Result: &ServiceLevelResult{
									Actual:   93.0,
									Delta:    -6.0,
									SloIsMet: false,
								},
							},
						},
					},
				},
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

			assert.Equal(t, got.Timestamp, tt.want.Timestamp, "Timestamp mismatch")
			assert.Equal(t, got.Services[0].ServiceLevels[0].Result, tt.want.Services[0].ServiceLevels[0].Result, "service level result mismatch")
			assert.Equal(t, got.Services[0].Dependencies, tt.want.Services[0].Dependencies, "service dependencies mismatch")
		})
	}
}
