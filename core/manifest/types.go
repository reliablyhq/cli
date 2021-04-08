package manifest

import (
	"fmt"

	"github.com/reliablyhq/cli/core"
)

type (
	Manifest struct {
		// App          *AppInfo          `yaml:"app" json:"app"`
		ServiceLevel []*Service `yaml:"slo" json:"slo"`
		Dependencies []string   `yaml:"dependencies" json:"dependencies"`
		// ServiceLevel *ServiceLevel     `yaml:"service_level,omitempty" json:"service_level,omitempty"`
		// CI           *ContinuousIntegrationInfo `yaml:"ci,omitempty" json:"ci,omitempty"`
		// Hosting      *Hosting          `yaml:"hosting,omitempty" json:"hosting,omitempty"`
		// IAC          *IAC              `yaml:"infrastructure_as_code,omitempty" json:"infrastructure_as_code,omitempty"`
		Tags map[string]string `yaml:"tags,omitempty" json:"tags,omitempty"`
	}

	AppInfo struct {
		Name       string `yaml:"name" json:"name"`
		Owner      string `yaml:"owner" json:"owner"`
		Repository string `yaml:"repo" json:"repo"`
	}

	Service struct {
		Name      string                `yaml:"name" json:"name"`
		Objective ServiceLevelObjective `yaml:"objective" json:"objective"`
		Resources []ServiceResource     `yaml:"resources" json:"resources"`
	}

	ServiceLevelObjective struct {
		Latency            core.Duration `yaml:"latency" json:"latency"`
		ErrorBudgetPercent float64       `yaml:"error_budget_percent" json:"error_budget_percent"`
	}

	ServiceResource struct {
		ID       string `yaml:"id" json:"id"`
		Provider string `yaml:"provider" json:"provider"`
	}

	ServiceLevel struct {
		Availability       float64       `yaml:"availability" json:"availability"`
		Latency            core.Duration `yaml:"latency" json:"latency"`
		ErrorBudgetPercent float64       `yaml:"error_budget_pc" json:"error_budget_pc"`
	}

	ContinuousIntegrationInfo struct {
		Type string `yaml:"type" json:"type"`
	}

	Hosting struct {
		Provider string `yaml:"provider" json:"provider"`
	}

	IAC struct {
		Type string `yame:"type" json:"type"`
		Root string `yaml:"root" json:"root"`
	}
)

// Validate - validate manifest
func (m *Manifest) Validate() error {
	// check slo names are duplicated
	names := make(map[string]struct{}, len(m.ServiceLevel))
	for _, s := range m.ServiceLevel {
		if _, exists := names[s.Name]; exists {
			return fmt.Errorf("duplicate SLO Name detected: [%s]", s.Name)
		}
		names[s.Name] = struct{}{}
	}

	return nil
}
