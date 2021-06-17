package utils

import "strings"

func GetShortVersion(version string) (string, bool) {
	supportedVersions := []string{
		"v1",
	}
	shortVersion := strings.Split(version, "/")
	for _, v := range supportedVersions {
		if v == shortVersion[len(shortVersion)-1] {
			return v, true
		}
	}
	return "", false
}
