package config

import (
	"os"
)

var (
	// The directory where the application config can be found
	ConfigDir = "~/.config/reliably"

	// The file that contains application config
	ConfigFile = ConfigDir + "/config.yaml"

	// The hostname of the reliably web services
	Hostname = "reliably.com"

	// The Entity Server hostname. This may be merged in to the Hostame variable later.
	EntityServerHost = Hostname

	// The organization info matching the RELIABLY_ORG ID/username used
	OverriddenOrg *OrgInfo
)

func init() {
	if h := os.Getenv(envReliablyHost); h != "" {
		Hostname = h
	}

	if h := os.Getenv(envReliablyEntityServerHost); h != "" {
		EntityServerHost = h
	}
}
