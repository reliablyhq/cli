package init

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reliablyhq/cli/core/cli/question"
	"github.com/reliablyhq/cli/core/manifest"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

var (
	manifestPath        string
	supportedExtensions = []string{".yaml", ".json"}
)

func NewCommand() *cobra.Command {
	cmd := cobra.Command{
		Use:   "init",
		Short: "initialise reliably",
		Long:  longCommandDescription(),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return validateFilePath()
		},
		RunE: runE,
	}

	cmd.Flags().StringVarP(&manifestPath, "manifest-file", "f", "reliably.yaml", "the location of the manifest file")

	return &cmd
}

func runE(_ *cobra.Command, args []string) error {
	scanner := bufio.NewScanner(os.Stdin)

	var m *manifest.Manifest
	if _, err := os.Stat(manifestPath); err == nil {
		if m, err = manifest.Load(manifestPath); err != nil {
			return err
		}
	} else {
		m = &manifest.Manifest{
			Service: &manifest.Service{},
		}
	}

	populateManifestInteractively(m, scanner)

	f, err := os.Create(manifestPath)
	if err != nil {
		return err
	}
	defer f.Close()

	if err := yaml.NewEncoder(f).Encode(&m); err != nil {
		return err
	}

	return nil
}

func validateFilePath() error {
	for _, ext := range supportedExtensions {
		if strings.HasSuffix(manifestPath, ext) {
			return nil
		}
	}

	return fmt.Errorf("manifest file must have one of the these extensions: %v", supportedExtensions)
}

func populateManifestInteractively(m *manifest.Manifest, scanner *bufio.Scanner) {
	if question.WithBoolAnswer(scanner, "Are you building something that will be provided to customers 'as a service'? (y/n)") {
		m.Service.Objective = manifest.ServiceLevelObjective{
			ErrorBudgetPercent: question.WithFloat64Answer(scanner, "What percentage of requests to your service is it ok to have fail? This will be your 'error budget'.", 0, 100),
			Latency:            question.WithDurationAnswer(scanner, "What is the maximum request-response latency you want from this service"),
		}
		m.Service.Resources = []manifest.ServiceResource{}

		do := question.WithBoolAnswer(scanner, "Do you want to add a service resource?")
		for do {
			provider := question.WithStringAnswer(scanner, "What is the name of the resource provider (e.g. aws, gcp, azure, etc)?")
			resourceID := question.WithStringAnswer(scanner, "What is the ID of the resource? This could be the AWS ARN, azure resource ID, etc.")

			m.Service.Resources = append(m.Service.Resources, manifest.ServiceResource{
				ID: fmt.Sprintf("%s/%s", provider, resourceID),
			})

			do = question.WithBoolAnswer(scanner, "Do you want to add another dependency?")
		}
	}

	if question.WithBoolAnswer(scanner, "Does your application have 'service level' dependencies? (y/n)") {
		deps := make([]string, 0)

		addMore := true
		for addMore {
			deps = append(deps, question.WithStringAnswer(scanner, "what is the name of the dependency?"))
			addMore = question.WithBoolAnswer(scanner, "Do you want to add another dependency? (y/n)")
		}

		m.Dependencies = deps
	} else {
		m.Dependencies = nil
	}
}

func longCommandDescription() string {
	return heredoc.Doc(`
		Initialise the reliably manifest.

		The manifest describes the operational contraints of the application, as well as some metadata about the app that allows users to reach out and communicate with the maintainer.

		Usage:
		1. reliably init: this method interactively creates a manifest file, asking you questions on the command line and adding your answers to the manifest file.
		2. reliably init -y: this method automatically creates an empty manifest file that you can manually complete later.
		3. realibly init -f <path>: this method works the same as reliably init, but allows you to specify the location of the file. This is useful if you use a multi-repo approach to source control.
		4. reliably init -f <path> -y: this method works the same as reliably init -y, but allows you to specify the location of the file. This is useful if you use a multi-repo approach to source control.
	`)
}
