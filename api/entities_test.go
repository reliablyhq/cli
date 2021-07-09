package api

import "testing"

func Test_requestPath(t *testing.T) {
	type args struct {
		org     string
		version string
		kind    string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "returns expected value",
			args: args{
				org:     "testing",
				version: "reliably/v1",
				kind:    "something",
			},
			want:    "entities/testing/reliably/v1/something",
			wantErr: false,
		},
		{
			name: "lowercases the args",
			args: args{
				org:     "TestIng", // except the org name
				version: "relIAbly/V1",
				kind:    "SOMEthing",
			},
			want:    "entities/TestIng/reliably/v1/something",
			wantErr: false,
		},
		{
			name: "returns an error is org is empty",
			args: args{
				org:     "",
				version: "b",
				kind:    "c",
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "returns an error is version is empty",
			args: args{
				org:     "a",
				version: "",
				kind:    "c",
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "returns an error is kind is empty",
			args: args{
				org:     "a",
				version: "b",
				kind:    "",
			},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := requestPath(tt.args.org, tt.args.version, tt.args.kind)
			if (err != nil) != tt.wantErr {
				t.Errorf("requestPath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("requestPath() = %v, want %v", got, tt.want)
			}
		})
	}
}
