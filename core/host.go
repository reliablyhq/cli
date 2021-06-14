package core

import (
	"fmt"
	"strings"
)

// RESTPrefix returns the Reliably API URL prefix
func RESTPrefix(hostname string) string {
	if strings.Contains(hostname, "127.0.0.1") ||
		strings.Contains(hostname, "localhost") {
		return fmt.Sprintf("http://%s/api/v1/", hostname) // unsecure HTTP for dev
	}
	return fmt.Sprintf("https://%s/api/v1/", hostname)
}

// BaseHttpUrl returns the scheme and net loc for URLs
func BaseHttpUrl(hostname string) string {
	if strings.Contains(hostname, "127.0.0.1") ||
		strings.Contains(hostname, "localhost") {
		return fmt.Sprintf("http://%s/", hostname) // unsecure HTTP for dev
	}
	return fmt.Sprintf("https://%s/", hostname)
}
