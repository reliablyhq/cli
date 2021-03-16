package manifest

func Empty() *Manifest {
	return &Manifest{
		ServiceLevel: &ServiceLevel{},
		CI:           &ContinuousIntegrationInfo{},
		Hosting:      &Hosting{},
		IAC:          &IAC{},
	}
}
