package advice

import (
	"reflect"
	"testing"
	"time"
)

func Test_calculateDowntime(t *testing.T) {
	type args struct {
		availability float32
	}
	tests := []struct {
		name string
		args args
		want *Downtime
	}{
		{
			name: "90% availability",
			args: args{
				availability: 90,
			},
			want: &Downtime{
				Day:     8640000000000 * time.Nanosecond,
				Week:    60480000000000 * time.Nanosecond,
				Month:   262980000000000 * time.Nanosecond,
				Quarter: 788832000000000 * time.Nanosecond,
				Year:    3156192000000000 * time.Nanosecond,
			},
		},
		{
			name: "99.95% availability",
			args: args{
				availability: 99.95,
			},
			want: &Downtime{
				Day:     43200000000 * time.Nanosecond,
				Week:    302400000000 * time.Nanosecond,
				Month:   1315200000000 * time.Nanosecond,
				Quarter: 3942000000000 * time.Nanosecond,
				Year:    15768000000000 * time.Nanosecond,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := calculateDowntime(tt.args.availability); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("calculateDowntime() = %v, want %v", got, tt.want)
			}
		})
	}
}
