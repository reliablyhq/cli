package entities

import (
	"bufio"
	"os"
	"strings"

	"gopkg.in/yaml.v2"
)

type Labels map[string]string

type Selector Labels

// TypeMeta describes an individual object in the entity API
// with strings representing the type of the object and its API schema version
type TypeMeta struct {
	// APIVersion defines the versioned schema of this representation of an object
	APIVersion string `json:"apiVersion,omitempty" yaml:"apiVersion,omitempty"`

	// Kind is a string value representing the REST resource this object represents.
	// In CamelCase.
	Kind string `json:"kind,omitempty" yaml:"kind,omitempty"`
}

// ObjectMeta is metadata that all entities must have
type Metadata struct {
	// Name of the object representation
	Name string `json:"-" yaml:"-"`

	// Map of string keys and values that can be used to identify an entity
	Labels map[string]string `json:"labels,omitempty" yaml:"labels,omitempty"`

	// List of map of string keys and values that can be used to link entities together
	// by identifying relationships
	// All string keys and values in the same map are mandatory to make a link (AND operand)
	// while any map of the list is necessary to make a link (OR operand)
	// example:
	// RelatedTo:
	//   - service:web
	//     region:us
	//   - service:web
	//     region:eu
	// In this example, the current object representation will be linked to
	// any other entity that is a web service and located in US or EU regions
	RelatedTo []map[string]string `json:"relatedTo,omitempty" yaml:"relatedTo,omitempty"`
}

// Entity is the interface of an entity object
// It can be used for type asserting
type Entity interface {
	Version() string
	Kind() string
}

// Manifest - a slice of Objectives
type Manifest []*Objective

func (m *Manifest) LoadFromFile(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}

	scanner := bufio.NewScanner(f)
	scanner.Split(splitManifest)
	for scanner.Scan() {
		var o Objective
		if err := yaml.Unmarshal(scanner.Bytes(), &o); err != nil {
			return err
		}

		*m = append(*m, &o)
	}
	return nil
}

func splitManifest(data []byte, atEOF bool) (advance int, token []byte, err error) {
	sep := "---\n"

	// Return nothing if at end of file and no data passed
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}

	if i := strings.Index(string(data), sep); i >= 0 {
		return i + len(sep), data[0:i], nil
	}

	// If at end of file with data return the data
	if atEOF {
		return len(data), data, nil
	}

	return
}
