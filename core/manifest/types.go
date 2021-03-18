package manifest

import "github.com/reliablyhq/cli/core"

type (
	Manifest struct {
		App          *AppInfo                   `yaml:"app" json:"app"`
		Onwer        string                     `yaml:"owner" json:"owner"`
		CI           *ContinuousIntegrationInfo `yaml:"ci,omitempty" json:"ci,omitempty"`
		ServiceLevel *ServiceLevel              `yaml:"service_level,omitempty" json:"service_level,omitempty"`
		Dependencies []*Dependency              `yaml:"dependencies,omitempty" json:"dependencies,omitempty"`
		Hosting      *Hosting                   `yaml:"hosting,omitempty" json:"hosting,omitempty"`
		IAC          *IAC                       `yaml:"infrastructure_as_code,omitempty" json:"infrastructure_as_code,omitempty"`
		Tags         map[string]string          `yaml:"tags,omitempty" json:"tags,omitempty"`
	}

	AppInfo struct {
		Name       string `yaml:"name" json:"name"`
		Owner      string `yaml:"owner" json:"owner"`
		Repository string `yaml:"repository" json:"repository"`
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
