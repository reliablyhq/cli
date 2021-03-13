package core

import (
	"os"

	"gopkg.in/yaml.v2"
)

const defaultManifestPath = "relably.yaml"

// Manifest that describes a Reliably applciation
type Manifest struct {
	Type string
}

func LoadManifest() (*Manifest, error) {
	path := getManifestPath()
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var m Manifest

	if err := yaml.NewDecoder(file).Decode(&m); err != nil {
		return nil, err
	}

	return &m, nil
}

func getManifestPath() string {
	return defaultManifestPath
}
