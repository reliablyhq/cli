package config

import (
	"os"

	"github.com/mitchellh/go-homedir"
	"gopkg.in/yaml.v3"
)

func resolveConfigFilePath() string {
	path, _ := homedir.Expand(ConfigFile)
	return path
}

func readConfigFile() (*Config, error) {
	p := resolveConfigFilePath()
	bytes, err := os.ReadFile(p)
	if err != nil {
		if os.IsNotExist(err) {
			return NewConfig(), nil
		} else {
			return nil, err
		}
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

	p := resolveConfigFilePath()
	if err := os.WriteFile(p, bytes, 0600); err != nil {
		return err
	}
	return nil
}
