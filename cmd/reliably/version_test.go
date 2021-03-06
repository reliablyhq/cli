package cmd

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRootCmdVersion(t *testing.T) {

	var out bytes.Buffer

	cmd := NewCmdRoot()
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs(strings.Split("--version", " "))

	fErr := cmd.Execute()
	output := out.String()

	assert.Equal(t, nil, fErr, "Unexpected error")
	assert.NotEqual(t, "", output, "Version message is missing from stdout")

}

func TestCmdVersion(t *testing.T) {

	rescueStdout := os.Stdout // keep backup of the real stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	var out bytes.Buffer

	cmd := NewCmdVersion()
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs(strings.Split("", " "))

	fErr := cmd.Execute()
	_ = out.String()

	// back to normal state
	w.Close()
	captured, _ := io.ReadAll(r)
	os.Stdout = rescueStdout // restoring the real stdout

	assert.Equal(t, nil, fErr, "Unexpected error")
	assert.NotEqual(t, "", captured, "Version message is missing from stdout")

}
func TestFormatVersion(t *testing.T) {

	prefix := "Reliably CLI version"

	tests := []struct {
		name string
		args []string
		want string
	}{
		{
			name: "dev version",
			args: []string{"DEV", "", "ab12cd34"},
			want: fmt.Sprintf("%s DEV (rev ab12cd34)\n", prefix),
		},
		{
			name: "semantic version",
			args: []string{"1.2.3", "", ""},
			want: fmt.Sprintf("%s 1.2.3\n", prefix),
		},
		{
			name: "semver with v prefix",
			args: []string{"v1.2.3", "", ""},
			want: fmt.Sprintf("%s 1.2.3\n", prefix),
		},
		{
			name: "version and build date",
			args: []string{"v1.2.3", "2021-03-04", ""},
			want: fmt.Sprintf("%s 1.2.3 (2021-03-04)\n", prefix),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := FormatVersion(tt.args[0], tt.args[1], tt.args[2])
			assert.Equal(t, tt.want, v, "Unexpected version")
		})
	}

}
