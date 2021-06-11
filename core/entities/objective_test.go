package entities

import (
	"encoding/json"
	"testing"
	"time"

	"gopkg.in/yaml.v2"

	"github.com/reliablyhq/cli/core"
)

func TestObjective(t *testing.T) {

	var sliSelect Labels = make(Labels, 0)

	slo := Objective{
		TypeMeta: TypeMeta{APIVersion: "v1", Kind: "Objective"},
		Metadata: Metadata{
			Name: "SLO for tests",
			Labels: map[string]string{
				"name": "SLO for tests",
				"env":  "test",
			},
			RelatedTo: []map[string]string{
				{"any": "entity"},
				{"more": "complex", "relation": "entity"},
			},
		},

		Spec: ObjectiveSpec{
			IndicatorSelector: Selector(sliSelect),
			ObjectivePercent:  90,
			Window:            core.Duration{Duration: time.Duration(24 * time.Hour)},
		},
	}

	t.Log(slo)

	y, _ := yaml.Marshal(slo)
	t.Log("YAML\n", string(y), "---")

	j, _ := json.Marshal(slo)
	t.Log("JSON\n", string(j), "---")

}
