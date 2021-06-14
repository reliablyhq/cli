package config

import "os"

const DefaultHostName = "reliably.com"

func GetHostname() string {
	if h := os.Getenv(envReliablyHost); h != "" {
		return h
	}

	return DefaultHostName
}
