package initAlpha

import (
	"os"
	"time"

	initCmd "github.com/reliablyhq/cli/cmd/reliably/slo/init"

	"github.com/reliablyhq/cli/api"
	"github.com/reliablyhq/cli/core"
	"github.com/reliablyhq/cli/core/entities"
	v "github.com/reliablyhq/cli/version"
)

func AlpaInitRun(opts *initCmd.InitOptions) error {
	var sliSelect entities.Labels = entities.Labels{
		"abc": "123",
	}

	slo := entities.Objective{
		TypeMeta: entities.TypeMeta{APIVersion: "v1", Kind: "Objective"},
		Metadata: entities.Metadata{
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

		Spec: entities.ObjectiveSpec{
			IndicatorSelector: entities.Selector(sliSelect),
			ObjectivePercent:  90,
			Window:            core.Duration{Duration: time.Duration(24 * time.Hour)},
		},
	}

	hostname := core.Hostname()
	entityHost := core.Hostname()
	if v.IsDevVersion() {
		if hostFromEnv := os.Getenv("RELIABLY_ENTITY_HOST"); hostFromEnv != "" {
			entityHost = hostFromEnv
		}
	}

	apiClient := api.NewClientFromHTTP(api.AuthHTTPClient(hostname))
	org, _ := api.CurrentUserOrganization(apiClient, hostname)

	return api.CreateEntity(apiClient, entityHost, org.Name, slo)

	//return errors.New("This is alpha version of the 'slo init' command")
}
