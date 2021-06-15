package config

type (
	Config struct {
		AuthInfo   map[string]AuthInfo `yaml:"auths"`
		CurrentOrg OrgInfo             `yaml:"currentOrg"`
	}

	AuthInfo struct {
		Token    string `yaml:"token"`
		Username string `yaml:"username"`
	}

	OrgInfo struct {
		Name string `yaml:"name"`
		ID   string `yaml:"id"`
	}
)
