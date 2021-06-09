package entities

import (
	"encoding/json"
	"testing"
	"time"

	"gopkg.in/yaml.v2"
)

func TestIndicator(t *testing.T) {

	to := time.Now()
	from := to.Add(-time.Hour * 24)

	sli := Indicator{
		TypeMeta: TypeMeta{APIVersion: "v1", Kind: "Indicator"},
		Metadata: Metadata{
			Labels: map[string]string{
				"category":   "latency",
				"percentile": "99",
				"latency":    "250ms",
			},
		},

		Spec: IndicatorSpec{
			From:    from,
			To:      to,
			Percent: 97.678,
		},
	}

	t.Log(sli)

	y, _ := yaml.Marshal(sli)
	t.Log("YAML\n", string(y), "---")

	j, _ := json.Marshal(sli)
	t.Log("JSON\n", string(j), "---")

}
