package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNestedMapLookup(t *testing.T) {

	var deep = `{
		"root": {
			"internal": {
				"int": 42,
				"string": "string",
				"float64": 6.022e9
			}
		}
	}`
	var m = make(map[string]interface{})
	json.Unmarshal([]byte(deep), &m)

	//val, err := NestedMapLookup(m, "root", "internal", "float64")

	type args struct {
		m  map[string]interface{}
		ks []string
	}
	tests := []struct {
		name    string
		args    args
		want    interface{}
		wantErr error
	}{
		{
			name:    "no keys",
			args:    args{m: m, ks: make([]string, 0, 0)},
			want:    nil,
			wantErr: errors.New("NestedMapLookup needs at least one key"),
		},
		{
			name:    "key not found",
			args:    args{m: m, ks: []string{"invalid"}},
			want:    nil,
			wantErr: fmt.Errorf("key not found; remaining keys: %v", []string{"invalid"}),
		},
		{
			name: "nested map has no-string as key",
			args: args{
				m: map[string]interface{}{
					"root": map[int]interface{}{
						0: false,
					},
				},
				ks: []string{"root", "fake"}, // use another key to recurse to nested map
			},
			want:    nil,
			wantErr: fmt.Errorf("malformed structure at %#v", map[int]interface{}{0: false}),
		},
		{
			name:    "root.internal.int",
			args:    args{m: m, ks: []string{"root", "internal", "int"}},
			want:    float64(42), // number loaded from json is float64
			wantErr: nil,
		},
		{
			name:    "root.internal.string",
			args:    args{m: m, ks: []string{"root", "internal", "string"}},
			want:    "string",
			wantErr: nil,
		},
		{
			name:    "root.internal.float64",
			args:    args{m: m, ks: []string{"root", "internal", "float64"}},
			want:    6.022 * math.Pow10(9),
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			val, err := NestedMapLookup(tt.args.m, tt.args.ks...)

			t.Log(val, err)
			t.Logf("%T, %T", val, err)

			if tt.wantErr != nil {
				if err == nil {
					t.Fatalf("no error in result, expected %v", tt.wantErr)
				} else if err.Error() != tt.wantErr.Error() {
					t.Fatalf("expected error %q, got %q", tt.wantErr.Error(), err.Error())
				}
				return
			}

			assert.Equal(t, nil, err, fmt.Sprintf("got error %v", err))
			assert.Equal(t, tt.want, val, "Map loook up did not return expected value")
		})
	}

}
