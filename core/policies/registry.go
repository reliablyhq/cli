package policies

type PolicyRegistry struct {
	registry   Store
	downloader Downloader
}

func (pr *PolicyRegistry) GetPolicy(id string) ([]byte, error) {
	exists := pr.registry.HasPolicy(id)

	if !exists {
		bs, err := pr.downloader.DownloadPolicy(id)
		if err != nil {
			return bs, err
		}

		err = pr.registry.UpsertPolicy(id, bs)
		if err != nil {
			// we were not able to save it in store,
			// but we can return the policy downloaded above
			return bs, nil
		}

		return bs, nil // no need to fetch it again from registry,
		//returns the same data that was inserted into the registry
	}

	bs, err := pr.registry.GetPolicy(id)
	if err != nil { // very unlickely as it's already been checked for existence
		return []byte{}, ErrPolicyNotFound
	}

	return bs, nil
}

func (pr *PolicyRegistry) ListPolicies() []string {
	policies, _ := pr.registry.ListPolicies()
	return policies
}

// NewRegistry returns a new instance of policy registry with default
// registry storage and policy downloader
// registry can be created with default options as r := NewRegistry()
// or with itw own custom options - ie specify a registry storage backend
// r := NewRegistry(MemStore)
func NewRegistry(options ...RegistryOption) *PolicyRegistry {

	r := &PolicyRegistry{
		registry:   NewMemStore(),
		downloader: &PolicyDownloader{},
	}

	for _, option := range options {
		option(r)
	}

	return r
}

type RegistryOption func(*PolicyRegistry)

var (
	MemStore RegistryOption = func(pr *PolicyRegistry) {
		pr.registry = NewMemStore()
	}
	FSStore = func(basedir string) RegistryOption {
		return func(pr *PolicyRegistry) {
			pr.registry = NewFSStore(basedir)
		}
	}
)
