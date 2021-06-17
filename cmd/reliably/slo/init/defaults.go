package init

import (
	"fmt"
	"strings"

	"github.com/reliablyhq/cli/core"
	"github.com/reliablyhq/cli/core/entities"
)

func generateDefaultSloName(o entities.Objective) string {
	var (
		name     string
		desc     string
		category string
	)

	if val, ok := o.Spec.IndicatorSelector["category"]; ok {
		category = val
	}

	switch category {
	case "latency":
		threshold := o.Spec.IndicatorSelector["latency_target"]

		if !strings.HasSuffix(threshold, "ms") {
			threshold = fmt.Sprintf("%sms", threshold)
		}

		desc = fmt.Sprintf("faster than %s", threshold)

	case "availability":
		desc = "successful"
	}

	name = fmt.Sprintf("%v%% of requests %s over last %s",
		o.Spec.ObjectivePercent,
		desc,
		core.HumanizeDuration(o.Spec.Window.Duration),
	)

	return name
}
