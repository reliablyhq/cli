package manifest

import (
	"github.com/reliablyhq/cli/core"
)

type (
	Manifest struct {
		App          *AppInfo          `yaml:"app" json:"app"`
		Service      *Service          `yaml:"service" json:"service"`
		Dependencies []string          `yaml:"dependencies" json:"dependencies"`
		Tags         map[string]string `yaml:"tags,omitempty" json:"tags,omitempty"`
		// ServiceLevel *ServiceLevel     `yaml:"service_level,omitempty" json:"service_level,omitempty"`
		// CI           *ContinuousIntegrationInfo `yaml:"ci,omitempty" json:"ci,omitempty"`
		// Hosting      *Hosting          `yaml:"hosting,omitempty" json:"hosting,omitempty"`
		// IAC          *IAC              `yaml:"infrastructure_as_code,omitempty" json:"infrastructure_as_code,omitempty"`
	}

	AppInfo struct {
		Name       string `yaml:"name" json:"name"`
		Owner      string `yaml:"owner" json:"owner"`
		Repository string `yaml:"repo" json:"repo"`
	}

	Service struct {
		Objective *ServiceLevelObjective `yaml:"objective" json:"objective"`
		Resources []*ServiceResource     `yaml:"resources" json:"resources"`
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
