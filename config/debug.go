package config

import "os"

func IsDebugMode() bool {
	_, ok := os.LookupEnv(envDebug)
	return ok
}
