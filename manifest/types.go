package manifest

import "time"

const defaultManifestPath = "reliably.yaml"

// Manifest that describes a Reliably applciation
type (
	AppType string

	Manifest struct {
		CI           *ContinuousIntegrationInfo `yaml:"ci,omitempty" json:"ci,omitempty"`
		ServiceLevel *ServiceLevel              `yaml:"service_level,omitempty" json:"service_level,omitempty"`
		Apps         []*AppInfo                 `yaml:"apps" json:"apps"`
		Dependencies []*Dependency              `yaml:"dependency" json:"dependency"`
		Hosting      *Hosting                   `yaml:"hosting" json:"hosting"`
		IAC          *IAC                       `yaml:"infrastructure_as_code" json:"infrastructure_as_code"`
	}

	ServiceLevel struct {
		Availability float32       `yaml:"availability" json:"availability"`
		Latency      time.Duration `yaml:"latency" json:"latency"`
		ErrorBudget  float32       `yaml:"error_budget" json:"error_budget"`
	}

	ContinuousIntegrationInfo struct {
		Type string `yaml:"type" json:"type"`
	}

	AppInfo struct {
		Name string `yame:"name" json:"name"`
		Root string `yaml:"root" json:"root"`
	}

	Dependency struct {
		Name          string `yaml:"name" json:"name"`
		Type          string `yaml:"type" json:"type"`
		RepositoryURL string `yaml:"repository_url" json:"repository_url"`
		StatusURL     string `yaml:"status_url" json:"status_url"`
	}

	Hosting struct {
		Provider       string            `yaml:"provider" json:"provider"`
		ConnectionInfo map[string]string `yaml:"connection_info" json:"connection_info"`
	}

	IAC struct {
		Type string `yame:"type" json:"type"`
		Root string `yaml:"root" json:"root"`
	}
)
