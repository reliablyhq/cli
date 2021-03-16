package manifest

func Empty() *Manifest {
	return &Manifest{
		ServiceLevel: &ServiceLevel{},
		CI:           &ContinuousIntegrationInfo{},
		Apps:         []*AppInfo{},
		Hosting:      &Hosting{},
		IAC:          &IAC{},
	}
}
