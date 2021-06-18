package config

// GetKnownHosts returns the list of hostnames defined in the config auths block
// If Token is provided by the env var, we use the overiddable host if not found in the config
func GetKnownHosts() ([]string, error) {
	cfg, err := readConfigFile()
	if err != nil {
		return nil, err
	}

	hosts := []string{}
	hasCurrentHost := false
	for h := range cfg.AuthInfo {
		hosts = append(hosts, h)
		if h == Hostname {
			hasCurrentHost = true
		}
	}

	if AuthTokenProvidedFromEnv() && !hasCurrentHost {
		hosts = append([]string{Hostname}, hosts...)
	}

	return hosts, nil
}
