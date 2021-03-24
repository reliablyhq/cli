package policies

import ()

// PolicyStorage defines the interface for policy module storage.
type PolicyStorage interface {
	ListPolicies() ([]string, error)
	GetPolicy(string) ([]byte, error)
	UpsertPolicy(string, []byte) error
	//DeletePolicy(string) error
	HasPolicy(string) bool
}

// Store defines the interface for any Policy Storage implementation
type Store interface {
	PolicyStorage
}
