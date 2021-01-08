package config

import (
	"fmt"
	"strings"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"

	"github.com/reliablyhq/cli/core"
)

var (
	// Viper is a global Viper object with custom options
	// - change the key delimiter to support dots in keys ie 'reliably.com'
	//   meaning the nested key lookup will be eg auths::reliably.com::token
	Viper = viper.NewWithOptions(viper.KeyDelimiter("::"))
)

// Hosts return the list of authentication hosts from config
// If Token is provided by the env var, but no host in the config
// we use the core overriddable hostname as single host
func Hosts() []string {
	hasDefault := false

	hosts := []string{}

	auths := Viper.GetStringMap("auths")
	for h := range auths {
		hosts = append(hosts, h)
		if h == core.Hostname() {
			hasDefault = true
		}
	}

	if core.AuthTokenProvidedFromEnv() && !hasDefault {
		hosts = append([]string{core.Hostname()}, hosts...)
	}

	return hosts

}

func GetAuthTokenWithSource(hostname string) (string, string, error) {

	if hostname == "" {
		hostname = core.Hostname()
	}

	if core.AuthTokenProvidedFromEnv() {
		token, env := core.AuthTokenFromEnv()
		return token, env, nil
	}

	tokenKey := fmt.Sprintf("auths::%s::token", hostname)
	token := Viper.GetString(tokenKey)
	if token != "" {
		// unexpand the absolute config paht to a ~ path
		home, _ := homedir.Dir()
		cfgFile := core.ConfigFile()
		cfgFile = strings.TrimPrefix(cfgFile, home)
		cfgFile = fmt.Sprintf("~%s", cfgFile)

		return token, cfgFile, nil
	}

	return "", "", fmt.Errorf("No token found for host %s", hostname)
}
