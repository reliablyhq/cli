package manifest

const defaultManifestPath = "reliably.yaml"

// Manifest that describes a Reliably applciation
type (
	AppType string

	Manifest struct {
		CI      *ContinuousIntegrationInfo `yaml:"name=ci,omitempty" json:"name=ci,omitempty"`
		Service *ServiceInfo               `yaml:"name=service,omitempty" json:"name=service,omitempty"`
		Apps    []*AppInfo                 `yaml:"name=app" json:"name=apps"`
	}

	ServiceInfo struct {
		DesiredAvailability float32 `yaml:"name=desired_availability" json:"name=desired_availability"`
	}

	ContinuousIntegrationInfo struct {
		Type string `yaml:"name=type" json:"name=type"`
	}

	AppInfo struct {
		Name string `yame:"name=name" json:"name=name"`
		Root string `yaml:"name=root" json:"name=root"`
	}
)
