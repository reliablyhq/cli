package api

import "testing"

func Test_requestPath(t *testing.T) {
	type args struct {
		version string
		kind    string
		org     string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "returns the expected value",
			args: args{
				org:     "some-org",
				version: "reliably.com/v1",
				kind:    "something",
			},
			want: "entities/some-org/reliably.com/v1/something",
		},
		{
			name: "converts uppercase stuff to lowercase",
			args: args{
				org:     "Some-Org",
				version: "Reliably.com/V1",
				kind:    "SomEThinG",
			},
			want: "entities/some-org/reliably.com/v1/something",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := requestPath(tt.args.version, tt.args.kind, tt.args.org); got != tt.want {
				t.Errorf("requestPath() = %v, want %v", got, tt.want)
			}
		})
	}
}
