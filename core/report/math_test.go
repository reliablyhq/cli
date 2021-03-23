package report

import (
	"testing"
)

func Test_sum(t *testing.T) {
	type args struct {
		array []float64
	}
	tests := []struct {
		name string
		args args
		want float64
	}{
		{
			name: "happy path",
			args: args{
				array: []float64{1, 2, 3, 4},
			},
			want: 10,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := sum(tt.args.array); got != tt.want {
				t.Errorf("sum() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_average(t *testing.T) {
	type args struct {
		array []float64
	}
	tests := []struct {
		name string
		args args
		want float64
	}{
		{
			name: "returns correct answer",
			args: args{
				array: []float64{10, 20, 30},
			},
			want: 20,
		},
		{
			name: "returns 0 if array is empty",
			args: args{
				array: []float64{},
			},
			want: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := average(tt.args.array); got != tt.want {
				t.Errorf("average() = %v, want %v", got, tt.want)
			}
		})
	}
}
