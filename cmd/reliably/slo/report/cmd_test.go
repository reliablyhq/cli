package report

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/google/shlex"
	"github.com/spf13/cobra"

	//"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/reliablyhq/cli/core/entities"
	"github.com/reliablyhq/cli/core/report"
	"github.com/stretchr/testify/assert"
)

func TestCommandOutputFlags(t *testing.T) {

	tests := []struct {
		name    string
		args    string
		wantErr string
	}{
		{
			name:    "single output format without path",
			args:    "-f json",
			wantErr: "",
		},
		{
			name:    "single output path without format (default)",
			args:    "-o report",
			wantErr: "",
		},
		{
			name:    "multiple output formats but no paths",
			args:    "-f yaml,json",
			wantErr: "Multiple output formats must be used in combination with multiple output path '--output o1,o2,...' flag",
		},
		{
			name:    "multiple output paths but no formats",
			args:    "-o o.yaml,o.json",
			wantErr: "Each output file specified with '--output' must have a format defined with '--format f1,f2,...'",
		},
		{
			name:    "not same number of values for formats & paths",
			args:    "-f yaml,json -o o.yaml,o.json,o.md",
			wantErr: "Flags '--format' and '--output' must have same number of values when combined",
		},
		{
			name:    "output format not supported",
			args:    "-f invalid",
			wantErr: fmt.Sprintf("Format 'invalid' is not valid. Use one of the supported formats: %v", supportedFormats),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			runF := func(opts *report.ReportOptions) error {
				//t.Log("overridden run function from test")
				//t.Log(*opts)
				return nil
			}

			cmd := NewCommand(runF)

			argv, err := shlex.Split(tt.args)
			require.NoError(t, err)
			cmd.SetArgs(argv)

			cmd.SetIn(&bytes.Buffer{})
			cmd.SetOut(ioutil.Discard)
			cmd.SetErr(ioutil.Discard)

			cmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
				//t.Log("disable auth check for command")
			}

			_, err = cmd.ExecuteC()
			if tt.wantErr != "" {
				require.EqualError(t, err, tt.wantErr)
				return
			} else {
				require.NoError(t, err)
			}

		})
	}

}

func TestMapToReports(t *testing.T) {

	var sliSelect entities.Labels = entities.Labels{
		"provider":       "gcp",
		"category":       "availability",
		"gcp_project_id": "abc123",
		"resource_id":    "projectid/google-cloud-load-balancers/loadbalancer-name",
	}

	objRes1 := entities.ObjectiveResultResponse{
		Metadata: entities.Metadata{
			Labels: map[string]string{
				"name":    "api-availability",
				"service": "example-api",
				"from":    "2021-06-13 12:07:57.081 +0000 UTC",
				"to":      "2021-06-14 12:07:57.081 +0000 UTC",
			},
		},
		Spec: entities.ObjectiveResultSpec{
			IndicatorSelector: entities.Selector(sliSelect),
			ObjectivePercent:  90,
			ActualPercent:     80,
			RemainingPercent:  -10,
		},
	}

	objRes2 := entities.ObjectiveResultResponse{
		Metadata: entities.Metadata{
			Labels: map[string]string{
				"name":    "api-latency",
				"service": "example-api",
				"from":    "2021-06-13 12:07:57.081 +0000 UTC",
				"to":      "2021-06-14 12:07:57.081 +0000 UTC",
			},
		},
		Spec: entities.ObjectiveResultSpec{
			IndicatorSelector: entities.Selector(sliSelect),
			ObjectivePercent:  80,
			ActualPercent:     90,
			RemainingPercent:  10,
		},
	}
	objResults := make([][]entities.ObjectiveResultResponse, 2)
	objResults[0] = []entities.ObjectiveResultResponse{objRes1}
	objResults[1] = []entities.ObjectiveResultResponse{objRes2}
	reports, err := report.MapToReports(objResults, 6, "v1")
	assert.NoError(t, err, "Error occurred in MapToReports")
	assert.Equal(
		t,
		reports[0].Services[0].Name,
		"example-api",
		"MapToReports incorrect service name mapping: example-api",
	)
	assert.Equal(
		t,
		reports[0].Services[0].ServiceLevels[0].Name,
		"api-availability",
		"MapToReports incorrect name mapping: api-availability",
	)
	assert.Equal(
		t,
		reports[0].Services[0].ServiceLevels[1].Name,
		"api-latency",
		"MapToReports incorrect name mapping: api-latency",
	)

}
