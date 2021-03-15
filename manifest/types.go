package manifest

const defaultManifestPath = "reliably.yaml"

// Manifest that describes a Reliably applciation
type (
	AppType string

	Manifest struct {
		CI      *ContinuousIntegrationInfo `yaml:"ci,omitempty" json:"ci,omitempty"`
		Service *ServiceInfo               `yaml:"service,omitempty" json:"service,omitempty"`
		Apps    []*AppInfo                 `yaml:"apps" json:"apps"`
	}

	ServiceInfo struct {
		DesiredAvailability float32 `yaml:"desired_availability" json:"desired_availability"`
	}

	ContinuousIntegrationInfo struct {
		Type string `yaml:"type" json:"type"`
	}

	AppInfo struct {
		Name string `yame:"name" json:"name"`
		Root string `yaml:"root" json:"root"`
	}
)
