package core

import (
	"os"
)

const (
	RELIABLY_TOKEN = "RELIABLY_TOKEN"
)

func AuthTokenFromEnv() (string, string) {
	return os.Getenv(RELIABLY_TOKEN), RELIABLY_TOKEN
}

func AuthTokenProvidedFromEnv() bool {
	return os.Getenv(RELIABLY_TOKEN) != ""
}
