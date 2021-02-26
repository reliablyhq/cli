package core

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func copyMap(originalMap map[string]interface{}) map[string]interface{} {
	newMap := make(map[string]interface{})
	for key, value := range originalMap {
		newMap[key] = value
	}
	return newMap
}

func TestUnmarshalSuggestionLevel(t *testing.T) {

	defaultS := map[string]interface{}{
		"rule_id":         "RULE-123",
		"rule_definition": "this is a dummy rule for tests",
		"details":         "",
		"file":            "/tmp/_tmp12345",
		"line":            0,
		"column":          0,
		"platform":        "test",
		"type":            "",
		"name":            "",
	}

	var lZeroValue Level

	tests := []struct {
		name       string
		suggestion func() string
		want       Level
		wantErr    bool
	}{
		{
			// this test has been used to allow suggestion that does not have
			// the level field in it, we rather ignore it than returning an error
			name: "missing level",
			suggestion: func() string {
				s := copyMap(defaultS)
				str, err := json.Marshal(s)
				if err != nil {
					return ""
				}
				return string(str)
			},
			want:    lZeroValue,
			wantErr: false,
		},
		{
			name: "valid level",
			suggestion: func() string {
				s := copyMap(defaultS)
				s["level"] = "warning"
				str, err := json.Marshal(s)
				if err != nil {
					return ""
				}
				return string(str)
			},
			want:    Warning,
			wantErr: false,
		},
		{
			name: "invalid level",
			suggestion: func() string {
				s := copyMap(defaultS)
				s["level"] = "unknown"
				str, err := json.Marshal(s)
				if err != nil {
					return ""
				}
				return string(str)
			},
			want:    0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			var s Suggestion
			err := json.Unmarshal([]byte(tt.suggestion()), &s)

			if tt.wantErr {
				assert.NotEqual(t, nil, err, "Expected error not returned in result")
				return
			}

			assert.Equal(t, nil, err, "Unexpected error")
			assert.Equal(t, s.Level, tt.want, "Suggestion has unexpected level value")

		})
	}

}
