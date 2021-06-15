package cmdutil

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/reliablyhq/cli/config"
	"github.com/reliablyhq/cli/core"
	"github.com/reliablyhq/cli/core/color"
)

// CheckAuth garanties a valid token is given to the CLI, for making
// authenticated calls, either from env var or config file
func CheckAuth() bool {
	if core.AuthTokenProvidedFromEnv() {
		return true
	}

	return config.GetTokenFor(config.Hostname) != ""
}

// PrintRequireAuthMsg displays formatted message that authentication
// to reliably is mandatory to continue with the CLI command
func PrintRequireAuthMsg() {
	// TODO - write this to stderr
	fmt.Println(color.Bold("Welcome to Reliably CLI!"))
	fmt.Println()
	fmt.Println("To authenticate, please run `reliably auth login`.")
}

func DisableAuthCheck(cmd *cobra.Command) {
	if cmd.Annotations == nil {
		cmd.Annotations = map[string]string{}
	}

	cmd.Annotations["skipAuthCheck"] = "true"
}

func IsAuthCheckEnabled(cmd *cobra.Command) bool {
	if !cmd.Runnable() {
		return false
	}
	for c := cmd; c.Parent() != nil; c = c.Parent() {
		if c.Annotations != nil && c.Annotations["skipAuthCheck"] == "true" {
			return false
		}
	}

	return true
}
