package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/MakeNowJust/heredoc/v2"
	homedir "github.com/mitchellh/go-homedir"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	v "github.com/reliablyhq/cli/version"
)

const (
	version   = v.Version
	workspace = ".reliably"
)

var (
	cfgFile string
	verbose bool

	rootCmd = &cobra.Command{
		Use:           "reliably <command> [flags]",
		Short:         "Reliably CLI",
		Long:          `The Reliably Command Line Interface (CLI).`,
		Version:       version,
		SilenceErrors: true, // quiet errors down stream
		SilenceUsage:  true, //silence usage when an error occurs
		Example: heredoc.Doc(`
			$ reliably discover`),
		Annotations: map[string]string{
			"help:feedback": heredoc.Doc(`

Feedback:
  You can provide with feedback or report an issue at https://github.com/reliablyhq/cli/issues/new
`),
		},
	}
)

// Execute the root command and exit with nonzero code in case of errors
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	createReliablyWorkspace()
	cobra.OnInitialize(initConfig)
	/*
		rootCmd.PersistentFlags().StringVar(
			&cfgFile, "config", "",
			"config file (default is $HOME/.reliably/config.yaml)")
	*/
	rootCmd.PersistentFlags().BoolVarP(
		&verbose, "verbose", "v", false, "verbose output")
	rootCmd.SetVersionTemplate(FormatVersion(version))
	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		if err := setUpVerboseLogLevel(verbose); err != nil {
			return err
		}
		return nil
	}
	// help will not be indicated as command but only flag
	rootCmd.SetHelpCommand(&cobra.Command{
		Use:    "no-help",
		Hidden: true,
	})

	// Uses a custom usage template - for adding feedback section to help -
	template := customUsageTemplate(rootCmd)
	rootCmd.SetUsageTemplate(template)
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".scanner" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".scanner")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}

//set the log level to debug if verbose mode is on
func setUpVerboseLogLevel(verbose bool) error {

	if verbose {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}

	return nil
}

func createReliablyWorkspace() {
	//Create dir output using above code
	if _, err := os.Stat(workspace); os.IsNotExist(err) {
		log.Debug("Create workspace '.reliably'")
		os.Mkdir(workspace, 0755)
	}

	// ensure policies cache is starting clean
	// so we remove before creating it
	// as removeAll removes path and any children it contains
	policiesFolder := filepath.Join(workspace, "policies")
	os.RemoveAll(policiesFolder)
	if _, err := os.Stat(policiesFolder); os.IsNotExist(err) {
		log.Debug(fmt.Sprintf("Create folder '%v'", policiesFolder))
		os.Mkdir(policiesFolder, 0755)
	}

	/*
		manifestsFolder := filepath.Join(workspace, "manifests")
		if _, err := os.Stat(manifestsFolder); os.IsNotExist(err) {
			log.Debug(fmt.Sprintf("Create folder '%v'", manifestsFolder))
			os.Mkdir(manifestsFolder, 0755)
		}
	*/

}

// CustomUsageTemplate returns usage template for the command.
// This is the default usage template from the command,
// with the additional help:feedback annotations, if specified
func customUsageTemplate(c *cobra.Command) string {

	tpl := c.UsageTemplate()

	if _, ok := c.Annotations["help:feedback"]; ok {
		feedback := c.Annotations["help:feedback"]
		return fmt.Sprintf("%s%s", tpl, feedback)
	}

	return tpl
}
