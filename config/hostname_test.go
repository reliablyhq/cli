package config

import (
	"os"
	"testing"
)

func TestGetHostname(t *testing.T) {
	tests := []struct {
		name    string
		want    string
		preFunc func()
	}{
		{
			name:    "returns the hostname from the environment",
			want:    "abc123",
			preFunc: func() { os.Setenv(envReliablyHost, "abc123") },
		},
		{
			name:    "returns default if env var has not been set",
			want:    DefaultHostName,
			preFunc: func() { os.Unsetenv(envReliablyHost) },
		},
		{
			name:    "returns default if env var value is an empty string",
			want:    DefaultHostName,
			preFunc: func() { os.Setenv(envReliablyHost, "") },
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.preFunc()
			if got := GetHostname(); got != tt.want {
				t.Errorf("GetHostname() = %v, want %v", got, tt.want)
			}
		})
	}
}
