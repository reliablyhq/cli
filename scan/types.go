package scan

import "encoding/json"

// Target - struct represents the entity/element to be
// scanned/evaluated
type Target struct {
	ResourceType string
	Platform     string
	Item         interface{}
	Metadata     map[string]string
	subgroups    []string
}

// Result - stores result of scan evaluation
type Result struct {
	Suggestions []string
	Violations  []Rule
}

// Rule - type used to unmarshal Rego Policy violations[rule]
type Rule struct {
	Level          uint32 `json:"level"`
	Message        string `json:"message"`
	RuleID         string `json:"ruleID"`
	RuleDefinition string `json:"ruleDef"`
}

func (r Rule) String() string {
	b, _ := json.MarshalIndent(&r, "", "  ")
	return string(b)
}

type policy struct {
	filepath string
	uri      string
}
