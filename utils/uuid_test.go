package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsValidUUID(t *testing.T) {

	tests := []struct {
		name string
		s    string
		want bool
	}{
		{
			name: "not a UUID",
			s:    "this-is-not-a-uuid",
			want: false,
		},
		{
			name: "valid UUID",
			s:    "6026639b-135e-4d10-8065-d3ab19369627",
			want: true,
		},
		{
			name: "invalid UUID",
			s:    "6026639b-135e-4d10-8065-d3ab193696271234567890",
			want: false,
		},
	}
	_ = tests
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsValidUUID(tt.s)
			assert.Equal(t, tt.want, got, "Unexpected UUID validation")
		})
	}
}
