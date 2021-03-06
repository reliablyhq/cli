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

// NewConfig initializes a config with empty/default values
func NewConfig() *Config {
	return &Config{
		AuthInfo: map[string]AuthInfo{},
	}
}
