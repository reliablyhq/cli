package config

import "os"

var (
	FilePath         = "~/.config/reliably/config.yaml"
	Hostname         = "reliably.com"
	EntityServerHost = Hostname
)

func init() {
	if h := os.Getenv(envReliablyHost); h != "" {
		Hostname = h
	}

	if h := os.Getenv(envReliablyEntityServerHost); h != "" {
		EntityServerHost = h
	}
}
