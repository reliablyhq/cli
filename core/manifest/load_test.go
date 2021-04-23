package manifest

import (
	"reflect"
	"testing"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/reliablyhq/cli/core"
	"github.com/stretchr/testify/assert"
)

func TestLoad(t *testing.T) {
	dummyReliablyYamlManifestPath := "../../tests/reliably.yaml"
	dummyReliablyJsonManifestPath := "../../tests/reliably.json"

	type args struct {
		path string
	}

	tests := []struct {
		name    string
		args    args
		want    *Manifest
		wantErr bool
	}{
		{
			name: "returns a manifest that matches a source yaml file",
			args: args{
				path: dummyReliablyYamlManifestPath,
			},
			want: &Manifest{
				Services: []*Service{
					&Service{
						Name: "Service A",
						ServiceLevels: []*ServiceLevel{
							&ServiceLevel{
								Name:      "Service A Availability",
								Type:      "availability",
								Objective: 99,
								Indicators: []ServiceLevelIndicator{
									{
										Provider: "aws",
										ID:       "arn1",
									},
									{
										Provider: "gcp",
										ID:       "uri",
									},
								},
							},
							&ServiceLevel{
								Name: "Service A Latency",
								Type: "latency",
								Criteria: LatencyCriteria{
									Threshold: core.Duration{Duration: 300 * time.Millisecond},
								},
								Objective: 99,
								Indicators: []ServiceLevelIndicator{
									{
										Provider: "aws",
										ID:       "arn2",
									},
								},
							},
						},
						//Dependencies: []string{},
					},
					&Service{
						Name: "Service B",
						ServiceLevels: []*ServiceLevel{
							&ServiceLevel{
								Name:      "Service B Availability",
								Type:      "availability",
								Objective: 99,
								Indicators: []ServiceLevelIndicator{
									{
										Provider: "aws",
										ID:       "arn3",
									},
								},
							},
						},
						//Dependencies: []string{},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "returns a manifest that matches a source json file",
			args: args{
				path: dummyReliablyJsonManifestPath,
			},
			want: &Manifest{
				Services: []*Service{
					&Service{
						Name: "Service A",
						ServiceLevels: []*ServiceLevel{
							&ServiceLevel{
								Name:      "Service A Availability",
								Type:      "availability",
								Objective: 99,
								Indicators: []ServiceLevelIndicator{
									{
										Provider: "aws",
										ID:       "arn1",
									},
									{
										Provider: "gcp",
										ID:       "uri",
									},
								},
							},
							&ServiceLevel{
								Name: "Service A Latency",
								Type: "latency",
								Criteria: LatencyCriteria{
									Threshold: core.Duration{Duration: 300 * time.Millisecond},
								},
								Objective: 99,
								Indicators: []ServiceLevelIndicator{
									{
										Provider: "aws",
										ID:       "arn2",
									},
								},
							},
						},
						//Dependencies: []string{},
					},
					&Service{
						Name: "Service B",
						ServiceLevels: []*ServiceLevel{
							&ServiceLevel{
								Name:      "Service B Availability",
								Type:      "availability",
								Objective: 99,
								Indicators: []ServiceLevelIndicator{
									{
										Provider: "aws",
										ID:       "arn3",
									},
								},
							},
						},
						//Dependencies: []string{},
					},
				},
			},
			wantErr: false,
		},
		{
			name:    "returns an error if the path is empty",
			args:    args{},
			want:    nil,
			wantErr: true,
		},
		{
			name: "returns an error if the file does not exist",
			args: args{
				path: "abc123",
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Load(tt.args.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("Load() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr == true {
				return
			}

			for serviceIndex, s := range tt.want.Services {
				if !reflect.DeepEqual(s.ServiceLevels, got.Services[serviceIndex].ServiceLevels) {
					t.Errorf("Wanted Service.ServiceLevels to be %v but was %v", s.ServiceLevels, got.Services[serviceIndex].ServiceLevels)
					return
				}

				if len(s.ServiceLevels) == len(got.Services[serviceIndex].ServiceLevels) {
					for i, sl := range tt.want.Services[serviceIndex].ServiceLevels {
						if sl.Name != got.Services[serviceIndex].ServiceLevels[i].Name {
							t.Errorf("%v != %v", sl.Name, got.Services[serviceIndex].ServiceLevels[i].Name)
							return
						}

						for i2, sli := range tt.want.Services[serviceIndex].ServiceLevels[i].Indicators {
							if sli.ID != got.Services[serviceIndex].ServiceLevels[i].Indicators[i2].ID {
								t.Errorf("%v != %v", sli.ID, got.Services[serviceIndex].ServiceLevels[i].Indicators[i2].ID)
								return
							}
						}
					}
				} else {
					t.Errorf("Wanted Service.Resources to be %v but was %v", len(s.ServiceLevels), len(got.Services[serviceIndex].ServiceLevels))
					return
				}

				if !reflect.DeepEqual(s.Dependencies, got.Services[serviceIndex].Dependencies) {
					t.Errorf("Wanted Dependencies to be %v but got %v", s.Dependencies, got.Services[serviceIndex].Dependencies)
					return
				}
			}
		})
	}
}

func TestManifestExampleCanBeLoaded(t *testing.T) {
	example := "../../examples/reliably.yaml"
	got, err := Load(example)

	assert.NoError(t, err, "Manifest example could not be loaded")
	assert.NotEqual(t, nil, got, "Manifest should not be empty")
	spew.Dump(got)
}
