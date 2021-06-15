package config

import (
	"os"
	"testing"
)

func TestIsDebugMode(t *testing.T) {
	tests := []struct {
		name    string
		want    bool
		preFunc func()
	}{
		{
			name:    "returns true if env var has a value",
			want:    true,
			preFunc: func() { os.Setenv(envDebug, "true") },
		},
		{
			name:    "returns true if env var has no value",
			want:    true,
			preFunc: func() { os.Setenv(envDebug, "") },
		},
		{
			name:    "returns false if env var has not been set",
			want:    false,
			preFunc: func() { os.Unsetenv(envDebug) },
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.preFunc()
			if got := IsDebugMode(); got != tt.want {
				t.Errorf("IsDebugMode() = %v, want %v", got, tt.want)
			}
		})
	}
}
