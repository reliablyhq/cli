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
		Use:     "init",
		Short:   "initialise the slo portion of the manifest",
		Long:    longCommandDescription(),
		Example: examples(),
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
			ServiceLevel: &manifest.Service{},
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
	m.ServiceLevel = &manifest.Service{}
	if question.WithBoolAnswer(scanner, "Are you building something that will be provided to customers 'as a service'?") {
		m.ServiceLevel.Objective = manifest.ServiceLevelObjective{
			ErrorBudgetPercent: question.WithFloat64Answer(scanner, "What percentage of requests to your service is it ok to have fail? This will be your 'error budget'.", 0, 100),
			Latency:            question.WithDurationAnswer(scanner, "What is the maximum request-response latency you want from this service?"),
		}

		m.ServiceLevel.Resources = []manifest.ServiceResource{}

		do := question.WithBoolAnswer(scanner, "Do you want to add a service resource?")
		for do {
			provider := question.WithStringAnswer(scanner, "What is the name of the resource provider (e.g. aws, gcp, azure, etc)?")
			resourceID := question.WithStringAnswer(scanner, "What is the ID of the resource? This could be the AWS ARN, azure resource ID, etc.")

			m.ServiceLevel.Resources = append(m.ServiceLevel.Resources, manifest.ServiceResource{
				ID: fmt.Sprintf("%s/%s", provider, resourceID),
			})

			do = question.WithBoolAnswer(scanner, "Do you want to add another dependency?")
		}
	}
}

func longCommandDescription() string {
	return heredoc.Doc(`
Initialise the reliably manifest.

The manifest describes the operational contraints of the application,
as well as some metadata about the app that allows users to reach out
and communicate with the maintainer.`)
}

func examples() string {
	return heredoc.Doc(`
$ reliably init:
  this method interactively creates a manifest file, asking you questions
  on the command line and adding your answers to the manifest file.

$ realibly init -f <path>:
  this method works the same as reliably init, but allows you to specify
  the location of the file. This is useful if you use a multi-repo approach
  to source control.`)
}
