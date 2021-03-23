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
				Dependencies: []*AppInfo{
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
				Dependencies: []*AppInfo{
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

			if !reflect.DeepEqual(tt.want, got) {
				t.Errorf("Wanted %v but got %v", tt.want.App, got.App)
			}
		})
	}
}
