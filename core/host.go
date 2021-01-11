package core

import (
	"fmt"
	"strings"
)

const defaultHostname = "reliably.com"

var hostname string

// DefaultHostname returns the host name of the default Reliably instance
func DefaultHostname() string {
	return defaultHostname
}

// Hostname returns the hostname,
// except it is overridable by the RELIABLY_HOST environment variable
func Hostname() string {
	if hostname != "" {
		return hostname
	}
	return defaultHostname
}

// SetHostname overrides the value returned from Hostname.
func SetHostname(newhost string) {
	hostname = newhost
}

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
