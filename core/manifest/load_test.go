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
					Name:               "some service",
					Latency:            core.Duration{Duration: 100 * time.Second},
					ErrorBudgetPercent: 2,
					Resources: []*ServiceResource{
						{
							ID: "some ID",
						},
					},
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
					Name:               "some service",
					Latency:            core.Duration{Duration: 100 * time.Millisecond},
					ErrorBudgetPercent: 2,
					Resources: []*ServiceResource{
						{
							ID: "some ID",
						},
					},
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
