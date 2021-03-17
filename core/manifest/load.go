package manifest

import (
	"encoding/json"
	"errors"
	"fmt"
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
	var decoder interface{ Decode(interface{}) error }

	ext := getExtensionFromPath(path)
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

func getExtensionFromPath(path string) string {
	i := strings.LastIndex(path, ".")
	return path[i:]
}
