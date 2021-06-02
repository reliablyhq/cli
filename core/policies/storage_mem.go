package policies

// store reprensents a structure for in-memory storage of policies
type store struct {
	policies map[string][]byte // raw policies
}

// NewMemStore returns an empty in-memory store.
func NewMemStore() Store {
	return &store{
		policies: map[string][]byte{},
	}
}

func (s *store) ListPolicies() ([]string, error) {
	keys := make([]string, 0, len(s.policies))
	for k := range s.policies {
		keys = append(keys, k)
	}
	return keys, nil
}

func (s *store) GetPolicy(id string) ([]byte, error) {
	bs, found := s.policies[id]
	if !found {
		return nil, ErrPolicyNotFound
		//return nil, fmt.Errorf("Policy '%s' not found", id)
	}
	return bs, nil
}

func (s *store) UpsertPolicy(id string, bs []byte) error {
	s.policies[id] = bs
	return nil
}

func (s *store) HasPolicy(id string) bool {
	ids, err := s.ListPolicies()
	if err != nil {
		return false
	}

	for _, pID := range ids {
		if pID == id {
			return true
		}
	}

	return false
}
