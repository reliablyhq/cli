package core

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v2"
)

const defaultManifestPath = "reliably.yaml"

// Manifest that describes a Reliably applciation
type (
	AppType string

	Manifest struct {
		Type AppType                   `yaml:"type",json:"type"`
		CI   ContinuousIntegrationInfo `yaml:"ci",json:""`
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
	var decoder interface{ Decode(interface{}) error }

	ext := getExtensionFromPath(p)
	switch ext {
	case ".yaml":
		{
			decoder = yaml.NewDecoder(file)
		}
	case ".json":
		{
			decoder = json.NewDecoder(file)
		}
	default:
		{
			return nil, fmt.Errorf("file type '%s' is not a supported manifest type", ext)
		}
	}

	if err := decoder.Decode(&m); err != nil {
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

func getExtensionFromPath(path string) string {
	i := strings.LastIndex(path, ".")
	return path[i:]
}
