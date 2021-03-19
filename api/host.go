package api

import "os"

const defaultApiServerHostUrl = "https://api.reliably.com"

func GetReliablyApiServerHostURL() string {
	if x := os.Getenv("RELIABLY_API_HOST_URL"); x != "" {
		return x
	}

	return defaultApiServerHostUrl
}
