package cmd

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/google/shlex"
	"github.com/reliablyhq/cli/core/iostreams"
	"github.com/stretchr/testify/assert"
)

func TestNewCmdWorkflow(t *testing.T) {

	tests := []struct {
		name    string
		args    string
		want    *WorkflowOptions
		wantErr string
	}{
		{
			name: "default usage",
			args: "",
			want: &WorkflowOptions{
				//Interactive: true, // as non TTY, we can not check that for now
				Stdout:   false,
				Platform: "",
			},
			wantErr: "",
		},
		{
			name: "set valid platform",
			args: "--platform github",
			want: &WorkflowOptions{
				Interactive: false,
				Stdout:      false,
				Platform:    "github",
			},
			wantErr: "",
		},
		{
			name:    "invalid platform",
			args:    "--platform unknown",
			want:    &WorkflowOptions{},
			wantErr: "Platform 'unknown' is not valid.",
		},
		{
			name: "redirect to stdout",
			args: "--stdout",
			want: &WorkflowOptions{
				Stdout: true,
			},
			wantErr: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			var out bytes.Buffer
			var gotOpts *WorkflowOptions

			args, err := shlex.Split(tt.args)
			assert.NoError(t, err)

			cmd := newCmdWorkflow(func(opts *WorkflowOptions) error {
				// capture opts from command constructor passed to running function
				gotOpts = opts
				return nil
			})
			cmd.SetOut(&out)
			cmd.SetErr(&out)
			cmd.SetArgs(args)

			fErr := cmd.Execute()
			_ = out.String()

			if tt.wantErr != "" {
				assert.Equal(t, true, strings.Contains(fErr.Error(), tt.wantErr),
					fmt.Sprintf("Invalid error: got '%s'; expected '%s'", fErr, tt.wantErr))
				return
			}

			assert.Equal(t, tt.want.Platform, gotOpts.Platform, "Invalid platform value")
			assert.Equal(t, tt.want.Stdout, gotOpts.Stdout, "Invalid stdout value")
			//assert.Equal(t, tt.want.Interactive, gotOpts.Interactive, "Invalid interactive value")
		})
	}

}

func TestWorkflowRunStdout(t *testing.T) {

	tests := supportedCIPlatforms

	for _, tt := range tests {
		t.Run(tt, func(t *testing.T) {
			io, _, out, _ := iostreams.Test() // creates io streams to buffers
			opts := &WorkflowOptions{
				IO:          io,
				Platform:    tt,
				Interactive: false,
				Stdout:      true,
			}

			err := workflowRun(opts)

			assert.NoError(t, err, "Unexpected error")
			assert.NotEqual(t, "", out, "Workflow has not been generated to stdout")

		})
	}
}
