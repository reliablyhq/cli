package entities

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
	Name string `json:"name,omitempty" yaml:"name,omitempty"`

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
