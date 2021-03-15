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
		DesiredAvailability float32 `yaml:"name=desired_availability,omitempty" json:"name=desired_availability,omitempty"`
	}

	ContinuousIntegrationInfo struct {
		Type string
	}

	AppInfo struct{}
)
