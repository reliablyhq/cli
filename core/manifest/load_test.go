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
				CI: &ContinuousIntegrationInfo{
					Type: "unit test ci",
				},
				ServiceLevel: &ServiceLevel{
					Availability:       75,
					Latency:            core.Duration{Duration: 200 * time.Millisecond},
					ErrorBudgetPercent: 0.5,
				},
				Dependencies: []*Dependency{
					{
						Name: "some service",
					},
					{
						Name: "some other service",
					},
				},
				Hosting: &Hosting{
					Provider: "some dummy provider",
				},
				IAC: &IAC{
					Type: "some IAC system",
					Root: "./abc",
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
				CI: &ContinuousIntegrationInfo{
					Type: "unit test ci",
				},
				ServiceLevel: &ServiceLevel{
					Availability:       75,
					Latency:            core.Duration{Duration: 200 * time.Millisecond},
					ErrorBudgetPercent: 0.5,
				},
				Dependencies: []*Dependency{
					{
						Name: "some service",
					},
					{
						Name: "some other service",
					},
				},
				Hosting: &Hosting{
					Provider: "some dummy provider",
				},
				IAC: &IAC{
					Type: "some IAC system",
					Root: "./abc",
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

			if tt.wantErr {
				return
			}

			if w, g := tt.want.App == nil, got.App == nil; w != g {
				t.Errorf("Wanted App == nil to be %v but was %v", w, g)
				return
			}

			if !reflect.DeepEqual(tt.want.App, got.App) {
				t.Errorf("Wanted App to be %v but was %v", tt.want.App, got.App)
				return
			}

			if w, g := tt.want.CI == nil, got.CI == nil; w != g {
				t.Errorf("Wanted CI == nil to be %v but was %v", w, g)
				return
			}

			if !reflect.DeepEqual(tt.want.CI, got.CI) {
				t.Errorf("Wanted CI to be %v but was %v", tt.want.CI, got.CI)
				return
			}

			if w, g := tt.want.Dependencies == nil, got.Dependencies == nil; w != g {
				t.Errorf("Wanted Dependencies == nil to be %v but was %v", w, g)
				return
			}

			if !reflect.DeepEqual(tt.want.Dependencies, got.Dependencies) {
				t.Errorf("Wanted Dependencies to be %v but was %v", tt.want.Dependencies, got.Dependencies)
				return
			}

			if w, g := tt.want.Hosting == nil, got.Hosting == nil; w != g {
				t.Errorf("Wanted Hosting == nil to be %v but was %v", w, g)
				return
			}

			if !reflect.DeepEqual(tt.want.Hosting, got.Hosting) {
				t.Errorf("Wanted Hosting to be %v but was %v", tt.want.Hosting, got.Hosting)
				return
			}

			if w, g := tt.want.IAC == nil, got.IAC == nil; w != g {
				t.Errorf("Wanted IAC == nil to be %v but was %v", w, g)
				return
			}

			if !reflect.DeepEqual(tt.want.IAC, got.IAC) {
				t.Errorf("Wanted IAC to be %v but was %v", tt.want.IAC, got.IAC)
				return
			}

			if w, g := tt.want.ServiceLevel == nil, got.ServiceLevel == nil; w != g {
				t.Errorf("Wanted ServiceLevel == nil to be %v but was %v", w, g)
				return
			}

			if !reflect.DeepEqual(tt.want.ServiceLevel, got.ServiceLevel) {
				t.Errorf("Wanted ServiceLevel to be %v but was %v", tt.want.ServiceLevel, got.ServiceLevel)
				return
			}

			if w, g := tt.want.Tags == nil, got.Tags == nil; w != g {
				t.Errorf("Wanted Tags == nil to be %v but was %v", w, g)
				return
			}

			if !reflect.DeepEqual(tt.want.Tags, got.Tags) {
				t.Errorf("Wanted Tags to be %v but was %v", tt.want.Tags, got.Tags)
				return
			}
		})
	}
}
