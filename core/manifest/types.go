package manifest

import (
	"fmt"

	"github.com/reliablyhq/cli/core"
)

type (
	Manifest struct {
		// App          *AppInfo          `yaml:"app" json:"app"`
		Services []*Service `yaml:"services" json:"services"`
		// Dependencies []string   `yaml:"dependencies" json:"dependencies"`
		// ServiceLevel *ServiceLevel     `yaml:"service_level,omitempty" json:"service_level,omitempty"`
		// CI           *ContinuousIntegrationInfo `yaml:"ci,omitempty" json:"ci,omitempty"`
		// Hosting      *Hosting          `yaml:"hosting,omitempty" json:"hosting,omitempty"`
		// IAC          *IAC              `yaml:"infrastructure_as_code,omitempty" json:"infrastructure_as_code,omitempty"`
		// Tags map[string]string `yaml:"tags,omitempty" json:"tags,omitempty"`
	}

	// AppInfo struct {
	// 	Name       string `yaml:"name" json:"name"`
	// 	Owner      string `yaml:"owner" json:"owner"`
	// 	Repository string `yaml:"repo" json:"repo"`
	// }

	Service struct {
		Name          string          `yaml:"name" json:"name"`
		ServiceLevels []*ServiceLevel `yaml:"service-levels" json:"service-levels"`
		Dependencies  []string        `yaml:"-" json:"-"`
	}

	ServiceLevel struct {
		Name     string      `yaml:"name" json:"name"`
		Type     string      `yaml:"type" json:"type"`
		Criteria interface{} `yaml:"criteria,omitempty" json:"criteria,omitempty"`
		//Threshold  core.Duration           `yaml:"threshold,omitempty" json:"threshold,omitempty"`
		Objective         float64                 `yaml:"slo" json:"slo"`
		Indicators        []ServiceLevelIndicator `yaml:"sli" json:"sli"`
		ObservationWindow core.Iso8601Duration    `yaml:"window" json:"window"`
	}

	Criteria struct {
	}

	LatencyCriteria struct {
		Threshold core.Duration `yaml:"threshold,omitempty" json:"threshold,omitempty"`
	}

	AvailabilityCriteria struct {
	}

	// ServiceLevelObjective struct {
	// 	Latency            core.Duration `yaml:"latency" json:"latency"`
	// 	ErrorBudgetPercent float64       `yaml:"error_budget_percent" json:"error_budget_percent"`
	// }

	ServiceLevelIndicator struct {
		ID       string `yaml:"id" json:"id"`
		Provider string `yaml:"provider" json:"provider"`
	}

	// ServiceLevel struct {
	// 	Availability       float64       `yaml:"availability" json:"availability"`
	// 	Latency            core.Duration `yaml:"latency" json:"latency"`
	// 	ErrorBudgetPercent float64       `yaml:"error_budget_pc" json:"error_budget_pc"`
	// }

	// ContinuousIntegrationInfo struct {
	// 	Type string `yaml:"type" json:"type"`
	// }

	// Hosting struct {
	// 	Provider string `yaml:"provider" json:"provider"`
	// }

	// IAC struct {
	// 	Type string `yame:"type" json:"type"`
	// 	Root string `yaml:"root" json:"root"`
	// }
)

// Validate - validate manifest
func (m *Manifest) Validate() error {
	// check slo names are duplicated
	servicesNames := make(map[string]struct{}, len(m.Services))
	for _, s := range m.Services {
		if _, exists := servicesNames[s.Name]; exists {
			return fmt.Errorf("Duplicate Service Name detected: [%s]", s.Name)
		}
		servicesNames[s.Name] = struct{}{}

		sloNames := make(map[string]struct{}, len(s.ServiceLevels))
		for _, sl := range s.ServiceLevels {
			if _, exists2 := sloNames[sl.Name]; exists2 {
				return fmt.Errorf("Duplicate Service Level Name detected in Service %s: [%s]", s.Name, sl.Name)
			}
			sloNames[sl.Name] = struct{}{}
		}
	}

	return nil
}
func (sl *ServiceLevel) UnmarshalYAML(unmarshal func(v interface{}) error) error {

	type Alias ServiceLevel

	// We unmarshal the ServiceLevel into the object
	var asl Alias
	if err := unmarshal(&asl); err != nil {
		fmt.Println("error unmarshall")
		return err
	}

	// assign the unmarshaled content to the current service level pointer
	*sl = ServiceLevel(asl)

	// Then we make a second unmarshalling
	// to create the right Criteria structure
	// depending on the service level type
	var criteria interface{}
	switch sl.Type {
	case "latency":

		lc := struct {
			Criteria LatencyCriteria `json:"criteria"`
		}{}

		if err := unmarshal(&lc); err == nil {
			criteria = lc.Criteria
		}

	case "availability":
	default:
	}

	// assign the parsed criteria to the current service level
	sl.Criteria = criteria

	return nil
}
