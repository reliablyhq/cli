package config

import "os"

var (
	FilePath = "~/.config/reliably/config.yaml"
	Hostname = "reliably.com"
)

func init() {
	if h := os.Getenv(envReliablyHost); h != "" {
		Hostname = h
	}
}
