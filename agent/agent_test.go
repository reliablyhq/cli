package agent

import (
	"testing"
	"time"

	"github.com/reliablyhq/cli/core/metrics"
	"github.com/reliablyhq/entity-server/server/types"
	"github.com/reliablyhq/entity-server/server/types/v1/service_level"
	"github.com/stretchr/testify/require"
)

var testJobObjective = &JobObjective{
	Objective: service_level.Objective{
		MetadataSpec: types.Metadata{
			Labels: types.Labels{
				"name": "solicito",
			},
		},
		Spec: service_level.ObjectiveSpec{
			IndicatorSelector: types.Selector{
				"category":       "latency",
				"percentile":     "90",
				"latency_target": "100ms",
				"latency":        "250",
			},
			ObjectivePercent: 99,
			Window:           types.Duration{Duration: time.Hour},
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
		}).IndicatorFunc(func(i *service_level.Indicator) error {
			labels := types.Labels(testJobObjective.Objective.Spec.IndicatorSelector)
			require.Equal(t,
				labels,
				i.MetadataSpec.Labels)

			require.Greater(t, i.Spec.Percent, float64(0))
			return nil
		}).Do()
	}()

	<-timeout
}
