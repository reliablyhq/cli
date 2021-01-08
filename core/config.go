package core

import (
	"path"

	"github.com/mitchellh/go-homedir"
)

type Config map[string]interface{}

func ConfigDir() string {
	dir, _ := homedir.Expand("~/.config/reliably")
	return dir
}

func ConfigFile() string {
	return path.Join(ConfigDir(), "config.yaml")
}
