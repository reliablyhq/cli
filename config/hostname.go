package config

import "os"

const DefaultHostName = "reliably.com"

func GetHostname() string {
	if h := os.Getenv("RELIABLY_HOST"); h != "" {
		return h
	}

	return DefaultHostName
}
