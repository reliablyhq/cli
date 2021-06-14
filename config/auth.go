package config

import (
	"os"
)

func GetTokenFor(hostname string) string {
	if token := os.Getenv(envReliablyToken); token != "" {
		return token
	}

	if info := getAuthInfoFromFile(hostname); info != nil {
		return info.Token
	} else {
		return ""
	}
}

func GetUsernameFor(hostname string) string {
	if info := getAuthInfoFromFile(hostname); info != nil {
		return info.Username
	} else {
		return ""
	}
}

func SetTokenForHostname(hostname, token string) error {
	cfg, err := readConfigFile()
	if err != nil {
		return err
	}

	cfg.AuthInfo[hostname] = AuthInfo{
		Token: token,
	}

	return writeConfigFile(cfg)
}

func SetUsernameForHostname(hostname, username string) error {
	cfg, err := readConfigFile()
	if err != nil {
		return err
	}

	cfg.AuthInfo[hostname] = AuthInfo{
		Username: username,
	}

	return writeConfigFile(cfg)
}

func SetAuthInfo(hostname string, data AuthInfo) error {
	cfg, err := readConfigFile()
	if err != nil {
		return err
	}

	cfg.AuthInfo[hostname] = data

	return writeConfigFile(cfg)
}

func DeleteAuthInfoForHostname(hostname string) error {
	cfg, err := readConfigFile()
	if err != nil {
		return err
	}

	delete(cfg.AuthInfo, hostname)

	return writeConfigFile(cfg)
}

func getAuthInfoFromFile(hostname string) *AuthInfo {
	cfg, err := readConfigFile()
	if err != nil {
		panic(err)
	}

	if info, ok := cfg.AuthInfo[hostname]; !ok {
		return nil
	} else {
		return &info
	}
}
