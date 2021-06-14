package config

func GetKnownHosts() ([]string, error) {
	cfg, err := readConfigFile()
	if err != nil {
		return nil, err
	}

	i := 0
	hosts := make([]string, len(cfg.AuthInfo))
	for key := range cfg.AuthInfo {
		hosts[i] = key
		i++
	}

	return hosts, nil
}
