package config

import "os"

func GetCurrentOrgInfo() (*OrgInfo, error) {
	cfg, err := readConfigFile()
	if err != nil {
		return nil, err
	}

	if orgID := os.Getenv("RELIABLY_ORG_ID"); orgID != "" {
		cfg.CurrentOrg.ID = orgID
	}

	if orgName := os.Getenv("RELIABLY_ORG_NAME"); orgName != "" {
		cfg.CurrentOrg.Name = orgName
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
