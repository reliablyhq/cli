package core

import (
	"os"

	"gopkg.in/yaml.v2"
)

const defaultManifestPath = "reliably.yaml"

// Manifest that describes a Reliably applciation
type (
	AppType string

	Manifest struct {
		Type AppType
		CI   ContinuousIntegrationInfo
	}

	ContinuousIntegrationInfo struct {
		Type string
	}
)

func LoadManifest(path string) (*Manifest, error) {
	p := getManifestPath(path)
	file, err := os.Open(p)
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

func getManifestPath(path string) string {
	s := defaultManifestPath

	if x := os.Getenv("RELIABLY_MANIFEST_PATH"); x != "" {
		s = x
	}

	if path != "" {
		s = path
	}

	return s
}
