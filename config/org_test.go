package config

import (
	"fmt"
	"log"
	"os"
	"reflect"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestGetCurrentOrgInfo(t *testing.T) {
	preFunc := func() error {
		cfg := Config{
			CurrentOrg: OrgInfo{
				Name: "some-org",
				ID:   "123abc",
			},
		}
		bytes, err := yaml.Marshal(&cfg)
		if err != nil {
			return err
		}
		defer log.Println("file has been written")
		return os.WriteFile(ConfigFile, bytes, 0755)
	}

	tests := []struct {
		name    string
		want    *OrgInfo
		wantErr bool
	}{
		{
			name: "returns value from the config file",
			want: &OrgInfo{
				Name: "some-org",
				ID:   "123abc",
			},
			wantErr: false,
		},
	}
	dir := ".temp"
	os.Mkdir(dir, 0755)
	defer os.RemoveAll(dir)

	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ConfigFile = fmt.Sprintf("%s/%d.yaml", dir, i)

			if err := preFunc(); err != nil {
				t.Error(err)
				return
			}

			got, err := GetCurrentOrgInfo()

			if (err != nil) != tt.wantErr {
				t.Errorf("GetCurrentOrgInfo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetCurrentOrgInfo() = %v, want %v", got, tt.want)
			}
		})
	}
}
