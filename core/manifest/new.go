package manifest

func New() *Manifest {
	return &Manifest{
		App:          &AppInfo{},
		CI:           &ContinuousIntegrationInfo{},
		ServiceLevel: &ServiceLevel{},
		Dependencies: []*Dependency{},
		Hosting:      &Hosting{},
		IAC:          &IAC{},
		Tags:         map[string]string{},
	}
}
