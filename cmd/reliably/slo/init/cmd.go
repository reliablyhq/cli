package init

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/reliablyhq/cli/core/cli/question"
	"github.com/reliablyhq/cli/core/manifest"
	"github.com/reliablyhq/cli/core/metrics"
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

	cmd.Flags().StringVarP(&manifestPath, "manifest-file", "f", manifest.DefaultManifestPath, "the location of the manifest file")

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
			Latency:            question.WithDurationAnswer(scanner, "What is the maximum request-response latency you want from this service (in milliseconds)?"),
		}

		m.ServiceLevel.Resources = []manifest.ServiceResource{}

		do := question.WithBoolAnswer(scanner, "Do you want to add a service resource?")
		if do {
			providers := []string{}
			for key := range metrics.ProviderFactories {
				providers = append(providers, key)
			}

			for do {
				provider := question.WithMultipleChoiceAnswer("What is the name of the resource provider?", providers...)
				id := getResourceIDForProvider(scanner, provider)

				m.ServiceLevel.Resources = append(m.ServiceLevel.Resources, manifest.ServiceResource{
					Provider: provider,
					ID:       id,
				})

				do = question.WithBoolAnswer(scanner, "Do you want to add another dependency?")
			}
		}
	}
}

func getResourceIDForProvider(scanner *bufio.Scanner, provider string) string {
	log.Print(provider)

	switch provider {
	case "aws":
		return question.WithStringAnswer(scanner, "What the ARN of the resource?")
	case "gcp":
		{
			projectID := question.WithStringAnswer(scanner, "What is the GCP project ID?")
			resourceType := "TODO: do this properly!"
			resourceName := question.WithStringAnswer(scanner, "What is the name of resource?")
			return fmt.Sprintf("%s/%s/%s", projectID, resourceType, resourceName)
		}
	default:
		return question.WithStringAnswer(scanner, "What is the ID of the resource? This could be the AWS ARN, azure resource ID, etc.")
	}
}
