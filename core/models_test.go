package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewLevel(t *testing.T) {

	tests := []struct {
		name    string
		level   string
		want    Level
		wantErr bool
	}{
		{
			name:    "level info",
			level:   "info",
			want:    Info,
			wantErr: false,
		},
		{
			name:    "level warning",
			level:   "warning",
			want:    Warning,
			wantErr: false,
		},
		{
			name:    "level error",
			level:   "error",
			want:    Error,
			wantErr: false,
		},
		{
			name:    "empty string",
			level:   "",
			want:    0,
			wantErr: true,
		},
		{
			name:    "invalid level",
			level:   "unknown",
			want:    0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			l, err := NewLevel(tt.level)

			if tt.wantErr {
				assert.NotEqual(t, nil, err, "Expected error not returned in result")
				return
			}

			assert.Equal(t, nil, err, "Unexpected error")
			assert.Equal(t, l, tt.want, "Unexpected level value")

		})
	}

}
