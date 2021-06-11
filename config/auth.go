package config

import (
	"os"
)

func GetTokenForHostname(hostname string) string {
	if token := os.Getenv("RELIABLY_TOKEN"); token != "" {
		return token
	}

	return getAuthInfoFromFile(hostname).Token
}

func GetUsernameForHostname(hostname string) string {
	if token := os.Getenv("RELIABLY_USERNAME"); token != "" {
		return token
	}

	return getAuthInfoFromFile(hostname).Username
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
	h := hostname
	if h == "" {
		h = os.Getenv("RELIABLY_HOST")
	}

	cfg, err := readConfigFile()
	if err != nil {
		panic(err)
	}

	return cfg.AuthInfo[h]
}
