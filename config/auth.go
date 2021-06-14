package config

import (
	"os"
)

func GetTokenForHostname(hostname string) string {
	if token := os.Getenv(envReliablyToken); token != "" {
		return token
	}

	return getAuthInfoFromFile(hostname).Token
}

func SetAuthInfoForHostname(hostname, token, username string) error {
	cfg, err := readConfigFile()
	if err != nil {
		return err
	}

	cfg.AuthInfo[hostname] = AuthInfo{
		Token:    token,
		Username: username,
	}

	return writeConfigFile(cfg)
}

func getAuthInfoFromFile(hostname string) AuthInfo {
	cfg, err := readConfigFile()
	if err != nil {
		panic(err)
	}

	if overridden_hostname := os.Getenv(envReliablyHost); overridden_hostname != "" {
		return cfg.AuthInfo[overridden_hostname]
	} else {
		return cfg.AuthInfo[hostname]
	}

}
