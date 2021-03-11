package scan

type EvalTarget struct {
	ResourceType string
	Platform     string
	Item         interface{}
	Metadata     map[string]string
}

type EvalResult struct {
	Suggestions []string
	Violations  []string
}

type policy struct {
	filepath string
}

func (p policy) String() string {
	return string(p)
}
