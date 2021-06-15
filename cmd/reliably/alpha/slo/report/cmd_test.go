package reportAlpha

import (
	"testing"

	"github.com/reliablyhq/cli/core/entities"
	"github.com/stretchr/testify/assert"
)

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
	reports, err := MapToReports(objResults, 6, "v1")
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
