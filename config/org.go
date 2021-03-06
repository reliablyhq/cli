package config

import ()

func GetCurrentOrgInfo() (*OrgInfo, error) {
	if OverriddenOrg != nil {
		return OverriddenOrg, nil
	}

	/*
		if orgID := os.Getenv(envReliablyOrg); orgID != "" {
			return &OrgInfo{
				ID: orgID,
			}, nil
		}
	*/

	cfg, err := readConfigFile()
	if err != nil {
		return nil, err
	}

	return &cfg.CurrentOrg, nil
}

func SetCurrentOrgInfo(name, ID string) error {
	cfg, err := readConfigFile()
	if err != nil {
		return err
	}

	cfg.CurrentOrg.Name = name
	cfg.CurrentOrg.ID = ID

	return writeConfigFile(cfg)
}
