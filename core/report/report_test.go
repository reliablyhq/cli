package report

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReportGetResult(t *testing.T) {

	r := Report{
		APIVersion: "test/v1",
		Services: []*Service{
			{
				Name: "another-service",
				ServiceLevels: []*ServiceLevel{
					{
						Name:      "my-first-slo",
						Objective: 99,
						Result: &ServiceLevelResult{
							SloIsMet: false,
						},
					},
				},
			},
			{
				Name: "test",
				ServiceLevels: []*ServiceLevel{
					{
						Name:      "my-second-slo",
						Objective: 99,
						Result: &ServiceLevelResult{
							SloIsMet: true,
						},
					},
					{
						Name:      "my-first-slo",
						Objective: 99,
						Result: &ServiceLevelResult{
							SloIsMet: false,
						},
					},
				},
			},
			{
				Name: "yet-another-service",
				ServiceLevels: []*ServiceLevel{
					{
						Name: "no-result",
					},
				},
			},
		},
	}

	tests := []struct {
		name    string
		svcName string
		sloName string
		want    *ServiceLevelResult
	}{
		{
			name:    "slo not found",
			svcName: "invalid",
			sloName: "invalid",
			want:    nil,
		},
		{
			name:    "slo without result",
			svcName: "yet-another-service",
			sloName: "no-result",
			want:    nil,
		},
		{
			name:    "slo with result",
			svcName: "test",
			sloName: "my-first-slo",
			want: &ServiceLevelResult{
				SloIsMet: false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := r.GetResult(tt.svcName, tt.sloName)
			if tt.want != nil && res != nil {
				assert.Equal(t, *tt.want, *res)
			} else {
				assert.Equal(t, tt.want, res)
			}

		})
	}

}
