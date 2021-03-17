package terraform

// PlanRepresentation document
type PlanRepresentation struct {
	FormatVersion   string                           `json:"format_version"`
	PriorState      *StateRepresentation             `json:"prior_state"`
	Configuration   *ConfigurationRepresentation     `json:"configuration"`
	PlannedValues   *ValuesRepresentation            `json:"planned_values"`
	Variables       map[string]map[string]string     `json:"variables"`
	ResourceChanges []*ResourceChangeRepresentation  `json:"resource_changes"`
	OutputChanges   map[string]*ChangeRepresentation `json:"output_changes"`
}

type StateRepresentation struct {
	TerraformVersion string                `json:"terraform_version"`
	Values           *ValuesRepresentation `json:"value"`
}

type ConfigurationRepresentation struct {
	ProviderConfigs *ProviderConfigRepresentation               `json:"provider_configs"`
	RootModule      *ModuleConfigRepresentation                 `json:"root_module"`
	ModuleCalls     map[string]*ModuleCallsConfigRepresentation `json:"module_calls"`
}

type ModuleConfigRepresentation struct {
	Outputs map[string]struct {
		Expression *ExpressionRepresentation `json:"expression"`
		Sensitive  bool                      `json:"sensitive"`
	} `json:"outputs"`

	Resources []struct {
		Address           string                         `json:"address"`
		Mode              string                         `json:"mode"`
		Type              string                         `json:"type"`
		Name              string                         `json:"name"`
		ProviderConfigKey string                         `json:"provider_config_key"`
		Expressions       *BlockExpressionRepresentation `json:"expressions"`
		SchemaVersion     int                            `json:"schema_version"`
		CountExpression   *ExpressionRepresentation      `json:"count_expression"`
		ForEachExpression *ExpressionRepresentation      `json:"for_each_expression"`
		Provisioners      []struct {
			Type        string                         `json:"type"`
			Expressions *BlockExpressionRepresentation `json:"expressions"`
		} `json:"provisioners"`
	} `json:"resources"`
}

type ModuleCallsConfigRepresentation struct {
	ResolvedSource    string                         `json:"resolved_source"`
	Expressions       *BlockExpressionRepresentation `json:"expressions"`
	CountExpression   *ExpressionRepresentation      `json:"count_expression"`
	ForEachExpression *ExpressionRepresentation      `json:"for_each_expression"`
	Module            *ModuleConfigRepresentation    `json:"module"`
}

type ValuesRepresentation struct {
	Outputs    map[string]*ValueRepresentation `json:"outputs"`
	RootModule *ModuleRepresentation           `json:"root_module"`
}

type ResourceChangeRepresentation struct {
	Address       string                `json:"address"`
	ModuleAddress string                `json:"module_address"`
	Mode          string                `json:"mode"`
	Type          string                `json:"type"`
	Name          string                `json:"name"`
	Index         int                   `json:"index"`
	Deposed       string                `json:"deposed"`
	Change        *ChangeRepresentation `json:"change"`
}

type ChangeRepresentation struct {
	Actions []string              `json:"actions"`
	Before  *ValuesRepresentation `json:"before"`
	After   *ValuesRepresentation `json:"after"`
}

type ValueRepresentation struct {
	Value     string `json:"string"`
	Sensitive bool   `json:"sensitive"`
}

type ModuleRepresentation struct {
	Resources    []*ResourceRepresentation    `json:"resources"`
	ChildModules []*ChildModuleRepresentation `json:"child_modules"`
}

type ResourceRepresentation struct {
	Address       string                 `json:"address"`
	Mode          string                 `json:"mode"`
	Type          string                 `json:"type"`
	Name          string                 `json:"name"`
	Index         int                    `json:"index"`
	ProviderName  string                 `json:"provider_name"`
	SchemaVersion int                    `json:"schema_version"`
	Values        map[string]interface{} `json:"values"`
}

type ChildModuleRepresentation struct {
	Address      string                       `json:"address"`
	Resources    []*ResourceRepresentation    `json:"resources"`
	ChildModules []*ChildModuleRepresentation `json:"child_modules"`
}

func (m ChildModuleRepresentation) ToModule() *ModuleRepresentation {
	return &ModuleRepresentation{
		Resources:    m.Resources,
		ChildModules: m.ChildModules,
	}
}

type ProviderConfigRepresentation struct {
	Name          string                         `json:"name"`
	Alias         string                         `json:"alias"`
	ModuleAddress string                         `json:"module_address"`
	Expressions   *BlockExpressionRepresentation `json:"expressions"`
}

type ExpressionRepresentation struct {
	ConstantValue string   `json:"constant_value"`
	References    []string `json:"references"`
}

type BlockExpressionRepresentation struct {
	AMI             *ExpressionRepresentation   `json:"ami"`
	InstanceType    *ExpressionRepresentation   `json:"instance_type"`
	RootBlockDevice *ExpressionRepresentation   `json:"root_block_device"`
	EBSBlockDevice  []*ExpressionRepresentation `json:"ebs_block_device"`
}

// TFResource -type used to create resource input for .tf file scanning
type TFResource struct {
	Label  string                 `json:"label"`
	Type   string                 `json:"type"`
	Values map[string]interface{} `json:"values"`
}
