package agent

import (
	"testing"
	"time"

	"github.com/reliablyhq/cli/core"
	"github.com/reliablyhq/cli/core/entities"
	"github.com/reliablyhq/cli/core/metrics"
	"github.com/stretchr/testify/require"
)

var testJobObjective = &JobObjective{
	Objective: entities.Objective{
		Metadata: entities.Metadata{
			Labels: entities.Labels{
				"name": "solicito",
			},
		},
		Spec: entities.ObjectiveSpec{
			IndicatorSelector: entities.Selector{
				"category":       "latency",
				"percentile":     "90",
				"latency_target": "100ms",
				"latency":        "250",
			},
			ObjectivePercent: 99,
			Window:           core.Duration{Duration: time.Hour},
		},
	},
	ResourceID: "alpha1-e3d83fa0/google-cloud-load-balancers/reliablyadvicealpha1",
}

func TestGetIndicatorFromObjective(t *testing.T) {
	i, err := GetIndicatorFromObjective(metrics.GCPProvider, testJobObjective)
	require.NoError(t, err)
	require.NotNil(t, i)
}

func TestAgent5Seconds(t *testing.T) {
	t.Parallel()
	// test Agent for 5 seconds
	timeout := time.After(5 * time.Second)
	go func() {
		objectives := []*JobObjective{testJobObjective}
		job := NewJob(5, objectives, metrics.GCPProvider)

		job.ErrorFunc(func(e *Error) {
			require.NoError(t, e)
		}).IndicatorFunc(func(i *entities.Indicator) error {
			require.Equal(t,
				entities.Labels(testJobObjective.Objective.Spec.IndicatorSelector),
				entities.Labels(i.Metadata.Labels))

			require.Greater(t, i.Spec.Percent, float64(0))
			return nil
		}).Do()
	}()

	<-timeout
}
