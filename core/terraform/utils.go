package terraform

import (
	"errors"

	"github.com/reliablyhq/cli/core"
)

// ExtractResourcesFromPlan analysis a plan and returns a collection of resources
func ExtractResourcesFromPlan(plan *PlanRepresentation) ([]*core.Resource, error) {
	if plan == nil {
		return nil, errors.New("plan is nil")
	}

	resources := make([]*core.Resource, len(plan.ResourceChanges))

	for i, resChange := range plan.ResourceChanges {
		resources[i] = &core.Resource{
			File: core.File{
				Filepath: "UNKNOWN",
			},
			StartingLine: 0,
			Platform:     Platform,
			Kind:         resChange.Type,
			Name:         resChange.Name,
			URI:          resChange.Address,
		}
	}

	return resources, nil
}
