package manifest

import (
	"fmt"
	"os"
	"reflect"
	"testing"
)

func TestLoad(t *testing.T) {
	type test struct {
		name    string
		want    *Manifest
		wantErr bool
		getPath func(*test) string
	}

	tests := []test{
		{
			name: "parses a yame file",
			want: &Manifest{
				Service: &ServiceInfo{
					DesiredAvailability: 95,
				},
				CI: &ContinuousIntegrationInfo{
					Type: "github actions",
				},
				Apps: []*AppInfo{
					{
						Name: "some-app",
						Root: ".",
					},
				},
			},
			wantErr: false,
			getPath: func(tx *test) string {
				dir := t.TempDir()
				if err := os.Chmod(dir, 777); err != nil {
					t.Fatal(err)
				}

				path := fmt.Sprintf("%s/reliably.yaml", dir)
				data := []byte("service:\n  desired_availability: 95\nci:\n  type: github actions\napps:\n- name: some-app\n  root: .")
				if err := os.WriteFile(path, data, 0); err != nil {
					t.Fatal(err)
				}
				return path
			},
		},
	}
	for _, tt := range tests {
		path := tt.getPath(&tt)

		t.Run(tt.name, func(t *testing.T) {
			got, err := Load(path)
			if (err != nil) != tt.wantErr {
				t.Errorf("Load() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Load() = %v, want %v", got, tt.want)
			}
		})
	}
}
