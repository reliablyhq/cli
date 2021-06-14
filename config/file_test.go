package config

import (
	"fmt"
	"os"
	"testing"
)

func Test_resolveConfigFilePath(t *testing.T) {
	tests := []struct {
		name    string
		want    string
		preFunc func()
	}{
		{
			name:    "resolves home folder",
			want:    fmt.Sprint(os.Getenv("HOME"), FilePath[1:]),
			preFunc: func() { FilePath = DefaultFilePath },
		},
		{
			name:    "resolves literal folder",
			want:    "/abc/123/xyz",
			preFunc: func() { FilePath = "/abc/123/xyz" },
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.preFunc()

			if got := resolveConfigFilePath(); got != tt.want {
				t.Errorf("resolveConfigFilePath() = %v, want %v", got, tt.want)
			}
		})
	}
}
