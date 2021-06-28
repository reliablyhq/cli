package utils

import "testing"

func TestGetShortVersion(t *testing.T) {

	type want struct {
		shortVersion string
		ok           bool
	}

	tests := []struct {
		name string
		arg  string
		want
	}{
		{
			name: "working standard api version",
			arg:  "reliably.com/v1",
			want: want{
				shortVersion: "v1",
				ok:           true,
			},
		},
		{
			name: "many slashes",
			arg:  "test/tests/v1",
			want: want{
				shortVersion: "v1",
				ok:           true,
			},
		},
		{
			name: "unsupported version",
			arg:  "reliably.com/v1000",
			want: want{
				shortVersion: "",
				ok:           false,
			},
		},
		{
			name: "double forward slash",
			arg:  "test//",
			want: want{
				shortVersion: "",
				ok:           false,
			},
		},
		{
			name: "empty string",
			arg:  "",
			want: want{
				shortVersion: "",
				ok:           false,
			},
		},
		{
			name: "already short version",
			arg:  "v1",
			want: want{
				shortVersion: "v1",
				ok:           true,
			},
		},
	}
	_ = tests
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotShortVersion, gotOk := GetShortVersion(tt.arg)
			got := want{
				shortVersion: gotShortVersion,
				ok:           gotOk,
			}
			if got != tt.want {
				t.Errorf("got: %v, want: %v", got, tt.want)
			}
		})
	}
}
