package scan

type Target struct {
	ResourceType string
	Platform     string
	Item         interface{}
	Metadata     map[string]string
	Result       Result
}

type Result struct {
	Suggestions []string
	Violations  []Rule
}

type Rule struct {
	Level          uint32 `json:"level"`
	Message        string `json:"message"`
	RuleID         string `json:"ruleID"`
	RuleDefinition string `json:"ruleDef"`
}

type policy struct {
	filepath string
	uri      string
}
