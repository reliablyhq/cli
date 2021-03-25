package manifest

func New() *Manifest {
	return &Manifest{
		App:     &AppInfo{},
		Service: &Service{},
		Tags:    map[string]string{},
	}
}
