package config

import (
	"io/fs"
	"os"

	"gopkg.in/yaml.v3"
)

const DefaultConfigFilePath = "~/.config/reliably/config.yaml"

var ConfigFilePath = DefaultConfigFilePath

func readConfigFile() (*Config, error) {
	bytes, err := os.ReadFile(ConfigFilePath)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := yaml.Unmarshal(bytes, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

func writeConfigFile(data *Config) error {
	bytes, err := yaml.Marshal(data)
	if err != nil {
		return err
	}

	if err := os.WriteFile(ConfigFilePath, bytes, fs.ModeAppend); err != nil {
		return err
	}
	return nil
}
