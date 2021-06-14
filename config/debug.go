package config

import "os"

func IsDebugMode() bool {
	return os.Getenv(envDebug) != ""
}
