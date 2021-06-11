package config

import "os"

func GetCurrentOrgInfo() (*OrgInfo, error) {
	if orgName := os.Getenv("RELIABLY_ORG"); orgName != "" {
		return &OrgInfo{
			Name: orgName,
		}, nil
	}

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
	cfg.CurrentOrg.ID = name

	return writeConfigFile(cfg)
}
