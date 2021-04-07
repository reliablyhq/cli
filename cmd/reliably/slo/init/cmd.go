package init

import (
	"fmt"
	"os"
	"strings"

	"github.com/reliablyhq/cli/core/cli/question"
	"github.com/reliablyhq/cli/core/color"
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

	cmd.Flags().StringVarP(&manifestPath, "manifest", "m", manifest.DefaultManifestPath, "the location of the manifest file")

	return &cmd
}

func runE(_ *cobra.Command, args []string) error {
	var m manifest.Manifest
	if _, err := os.Stat(manifestPath); err == nil {
		if !question.WithBoolAnswer(fmt.Sprintf("Existing manifest detected (%s); Do you want to overwrite it?", manifestPath)) {
			return nil
		}
	}

	populateManifestInteractively(&m)

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

func populateManifestInteractively(m *manifest.Manifest) {
	var s manifest.Service
	if question.WithBoolAnswer("Are you building something that will be provided to customers 'as a service'?") {
		s.Objective = manifest.ServiceLevelObjective{
			ErrorBudgetPercent: question.WithFloat64Answer("What percentage of requests to your service is it ok to have fail? This will be your 'error budget'.", 0, 100),
			Latency:            question.WithDurationAnswer("What is the maximum request-response latency you want from this service (in milliseconds)?"),
		}

		s.Resources = []manifest.ServiceResource{}

		do := question.WithBoolAnswer("Do you want to add a service resource?")
		if do {
			providers := []string{}
			for key := range metrics.ProviderFactories {
				providers = append(providers, key)
			}

			for do {
				provider := question.WithSingleChoiceAnswer("What is the name of the resource provider?", providers...)
				id := getResourceIDForProvider(provider)

				s.Resources = append(s.Resources, manifest.ServiceResource{
					Provider: provider,
					ID:       id,
				})

				do = question.WithBoolAnswer("Do you want to add another dependency?")
			}
		}

		s.Name = question.WithStringAnswer("SLO/Service name?")
	}
	m.ServiceLevel = append(m.ServiceLevel, &s)
	fmt.Println(color.Green(fmt.Sprintf("SLO/Service (%s) added", s.Name)))
	if question.WithBoolAnswer("Do you want to add another SLO?") {
		populateManifestInteractively(m)
	}
}

func getResourceIDForProvider(provider string) string {
	switch provider {
	case "aws":
		return question.WithStringAnswer("What is the ARN of the resource?")
	case "gcp":
		{
			projectID := question.WithStringAnswer("What is the GCP project ID?")
			resourceType := question.WithSingleChoiceAnswer("What is the 'type' of the resource?", "Google Cloud Load Balancers")
			resourceName := question.WithStringAnswer("What is the name of resource?")
			return fmt.Sprintf("%s/%s/%s", projectID, resourceType, resourceName)
		}
	default:
		return question.WithStringAnswer("What is the ID of the resource? This could be the AWS ARN, azure resource ID, etc.")
	}
}
