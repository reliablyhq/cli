package report

import (
	"testing"
	"time"
)

func TestIsoTimeParse(t *testing.T) {

	expectedTime, _ := time.Parse("2006-01-02 15:04:05.999999999 -0700 MST", "2021-07-13 10:01:01.000001 +0000 UTC")
	secondExpectedTime, _ := time.Parse("2006-01-02 15:04:05.999999999 -0700 MST", "2021-07-13 10:01:01.1 +0000 UTC")

	tests := []struct {
		name string
		arg  string
		want time.Time
	}{
		{
			name: "Format from time.String()",
			arg:  "2021-07-13 10:01:01.000001 +0000 UTC",
			want: expectedTime,
		},
		{
			name: "Postman's isotimestamp",
			arg:  "2021-07-13T10:01:01.000001Z",
			want: expectedTime,
		},
		{
			name: "Format from time.String() - nano",
			arg:  "2021-07-13 10:01:01.000001000 +0000 UTC",
			want: expectedTime,
		},
		{
			name: "Postman's isotimestamp - 1digits",
			arg:  "2021-07-13T10:01:01.1Z",
			want: secondExpectedTime,
		},
		{
			name: "Postman's isotimestamp - 3digits",
			arg:  "2021-07-13T10:01:01.100Z",
			want: secondExpectedTime,
		},
		{
			name: "Format from time.String() - 2digits",
			arg:  "2021-07-13 10:01:01.10 +0000 UTC",
			want: secondExpectedTime,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := isoTimeParse(test.arg)
			if err != nil {
				t.Errorf("error whilst parsing: %v", test.arg)
			}
			if got != test.want {
				t.Errorf("got: %v, want: %v", got, test.want)
			}
		})

	}

}
