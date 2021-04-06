package manifest

import (
	"errors"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"

	"gopkg.in/yaml.v2"
)

const DefaultManifestPath = "reliably.yaml"

func Load(path string) (*Manifest, error) {
	if path == "" {
		return nil, errors.New("path is empty")
	}

	log.Debug("Loading manifest at ", path)

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

func getExtensionFromPath(path string) string {
	i := strings.LastIndex(path, ".")
	return path[i:]
}
