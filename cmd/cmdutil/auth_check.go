package cmdutil

import (
	"fmt"

	"github.com/reliablyhq/cli/core"
	"github.com/reliablyhq/cli/core/color"
	"github.com/reliablyhq/cli/core/config"
)

// CheckAuth garanties a valid token is given to the CLI, for making
// authenticated calls, either from env var or config file
func CheckAuth() bool {
	if core.AuthTokenProvidedFromEnv() {
		return true
	}

	token, _, err := config.GetAuthTokenWithSource(core.Hostname())
	if err != nil {
		return false
	}
	if token != "" {
		return true
	}

	return false
}

// PrintRequireAuthMsg displays formatted message that authentication
// to reliably is mandatory to continue with the CLI command
func PrintRequireAuthMsg() {
	// TODO - write this to stderr
	fmt.Println(color.Bold("Welcome to Reliably CLI!"))
	fmt.Println()
	fmt.Println("To authenticate, please run `reliably auth login`.")
}
