package manifest

import "github.com/reliablyhq/cli/core"

type (
	Manifest struct {
		ApplicationName string                     `yaml:"application_name" json:"application_name"`
		CI              *ContinuousIntegrationInfo `yaml:"ci,omitempty" json:"ci,omitempty"`
		ServiceLevel    *ServiceLevel              `yaml:"service_level,omitempty" json:"service_level,omitempty"`
		Dependencies    []*Dependency              `yaml:"dependency" json:"dependency"`
		Hosting         *Hosting                   `yaml:"hosting,omitempty" json:"hosting,omitempty"`
		IAC             *IAC                       `yaml:"infrastructure_as_code,omitempty" json:"infrastructure_as_code,omitempty"`
		Tags            map[string]string          `yaml:"tags" json:"tags"`
	}

	ServiceLevel struct {
		Availability       float64       `yaml:"availability" json:"availability"`
		Latency            core.Duration `yaml:"latency" json:"latency"`
		ErrorBudgetPercent float64       `yaml:"error_budget_pc" json:"error_budget_pc"`
	}

	ContinuousIntegrationInfo struct {
		Type string `yaml:"type" json:"type"`
	}

	Dependency struct {
		Name string `yaml:"name" json:"name"`
	}

	Hosting struct {
		Provider string `yaml:"provider" json:"provider"`
	}

	IAC struct {
		Type string `yame:"type" json:"type"`
		Root string `yaml:"root" json:"root"`
	}
)
