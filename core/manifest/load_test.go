package manifest

import (
	"reflect"
	"testing"
	"time"

	"github.com/reliablyhq/cli/core"
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
				App: &AppInfo{
					Name:       "unit test app",
					Owner:      "unit test owner",
					Repository: "github.com/reliablyhq/cli",
				},
				Service: &Service{
					Latency:            core.Duration{Duration: 100 * time.Millisecond},
					ErrorBudgetPercent: 0.5,
					Resources:          []*ServiceResource{},
				},
				Dependencies: []string{
					"some service",
					"some other service",
				},
				Tags: map[string]string{
					"team":   "abc",
					"domain": "xyz",
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
				App: &AppInfo{
					Name:       "unit test app",
					Owner:      "unit test owner",
					Repository: "github.com/reliablyhq/cli",
				},
				Service: &Service{
					Latency:            core.Duration{Duration: 100 * time.Millisecond},
					ErrorBudgetPercent: 0.5,
					Resources: []*ServiceResource{
						{
							Provider: "abc",
							ID:       "123",
						},
						{
							Provider: "xyz",
							ID:       "456",
						},
					},
				},
				Dependencies: []string{
					"some service",
					"some other service",
				},
				Tags: map[string]string{
					"team":   "abc",
					"domain": "xyz",
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

			if !reflect.DeepEqual(tt.want.App, got.App) {
				t.Errorf("Wanted App to be %v but got %v", tt.want.App, got.App)
				return
			}

			if tt.want.Service.ErrorBudgetPercent != got.Service.ErrorBudgetPercent {
				t.Errorf("Wanted Service.ErrorBudgetPercent to be %v but was %v", tt.want.Service.ErrorBudgetPercent, got.Service.ErrorBudgetPercent)
				return
			}

			if tt.want.Service.Latency != got.Service.Latency {
				t.Errorf("Wanted Service.Latency to be %v but was %v", tt.want.Service.Latency, got.Service.Latency)
				return
			}

			if len(tt.want.Service.Resources) == len(got.Service.Resources) {
				for i, r := range tt.want.Service.Resources {
					if r.ID != got.Service.Resources[i].ID {
						t.Errorf("%v != %v", r.ID, got.Service.Resources[i].ID)
						return
					}

					if r.Provider != got.Service.Resources[i].Provider {
						t.Errorf("%v != %v", r.Provider, got.Service.Resources[i].Provider)
						return
					}
				}
			} else {
				t.Errorf("Wanted Service.Resources to be %v but was %v", len(tt.want.Service.Resources), len(got.Service.Resources))
				return
			}

			if !reflect.DeepEqual(tt.want.Dependencies, got.Dependencies) {
				t.Errorf("Wanted Dependencies to be %v but got %v", tt.want.Dependencies, got.Dependencies)
				return
			}

			if !reflect.DeepEqual(tt.want.Tags, got.Tags) {
				t.Errorf("Wanted Tags to be %v but got %v", tt.want.Tags, got.Tags)
				return
			}
		})
	}
}
