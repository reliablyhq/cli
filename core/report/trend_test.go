package report

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetSloTrend(t *testing.T) {

	reports := []Report{
		{
			APIVersion: "test/v1",
			Services: []*Service{
				{
					Name: "test",
					ServiceLevels: []*ServiceLevel{
						{
							Name:      "my-first-slo",
							Objective: 99,
							Result:    nil,
						},
					},
				},
			},
		},

		{
			APIVersion: "test/v1",
			Services: []*Service{
				{
					Name: "test",
					ServiceLevels: []*ServiceLevel{
						{
							Name:      "my-first-slo",
							Objective: 99,
							Result:    nil,
						},
					},
				},
			},
		},

		{
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
							Name:      "my-first-slo",
							Objective: 99,
							Result: &ServiceLevelResult{
								SloIsMet: true,
							},
						},
					},
				},
			},
		},

		{
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
			},
		},
	}

	var trends []bool

	trends = GetSLOTrend("dummy", "", reports)
	assert.Equal(t, []bool{}, trends)

	trends = GetSLOTrend("test", "my-invalid-slo", reports)
	assert.Equal(t, []bool{}, trends)

	trends = GetSLOTrend("test", "my-first-slo", reports)
	assert.Equal(t, []bool{true, false}, trends)

}
