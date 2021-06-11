package agent

import (
	"testing"
	"time"

	"github.com/reliablyhq/cli/core"
	"github.com/reliablyhq/cli/core/entities"
	"github.com/reliablyhq/cli/core/metrics"
	"github.com/stretchr/testify/require"
)

var testJobObjective = &entities.Objective{
	Metadata: entities.Metadata{
		Labels: entities.Labels{
			"name": "solicito",
		},
	},
	Spec: entities.ObjectiveSpec{
		IndicatorSelector: entities.Selector{
			"category":              "latency",
			"percentile":            "90",
			"latency_target":        "100ms",
			"latency":               "250",
			"gcp_loadbalancer_name": "staging-lb",
			"gcp_project_id":        "alpha1-e3d83fa0",
		},
		ObjectivePercent: 99,
		Window:           core.Duration{Duration: time.Hour},
	},
}

// 	ResourceID: "alpha1-e3d83fa0/google-cloud-load-balancers/reliablyadvicealpha1",
// }

func TestGetIndicatorFromObjective(t *testing.T) {
	t.Skip("skipping...")
	i, err := getIndicatorFromObjective(testJobObjective)
	require.NoError(t, err)
	require.NotNil(t, i)
}

func TestAgent5Seconds(t *testing.T) {
	t.Skip("skipping...")
	t.Parallel()
	// test Agent for 5 seconds
	timeout := time.After(5 * time.Second)
	go func() {
		objectives := []*entities.Objective{testJobObjective}
		job := NewJob(5, objectives, metrics.GCPProvider)

		job.ErrorFunc(func(e *Error) {
			require.NoError(t, e)
		}).IndicatorFunc(func(i *entities.Indicator) error {
			require.Equal(t,
				entities.Labels(testJobObjective.Spec.IndicatorSelector),
				entities.Labels(i.Metadata.Labels))

			require.Greater(t, i.Spec.Percent, float64(0))
			return nil
		}).Do()
	}()

	<-timeout
}
