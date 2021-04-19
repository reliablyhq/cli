package init

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/reliablyhq/cli/core/cli/question"
	"github.com/reliablyhq/cli/core/color"
	"github.com/reliablyhq/cli/core/manifest"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

var (
	manifestPath        string
	supportedExtensions = []string{".yaml", ".json"}
	googleResourceTypes = []string{"Google Cloud Load Balancers"}
	providersMap        = map[string]string{
		"Amazon Web Services":   "aws",
		"Google Cloud Platform": "gcp",
	}
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

	// validate
	if err := m.Validate(); err != nil {
		return err
	}

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

	// s.Type = manifest.Service.Type{}

	s.Name = question.WithStringAnswer("What is the name of the service you want to declare SLOs for?")

	declareSLOForService(&s)

	m.Services = append(m.Services, &s)
	fmt.Println(color.Green(fmt.Sprintf("Service '%s' added", s.Name)))

	if question.WithBoolAnswer("Do you want to add another Service?") {
		populateManifestInteractively(m)
	}

}

func declareSLOForService(s *manifest.Service) {
	var sl manifest.ServiceLevel

	slType := question.WithSingleChoiceAnswer("What type of SLO do you want to declare?", "Availability", "Latency")
	sl.Type = sanitizeString(slType)

	if sl.Type == "latency" {
		threshold := question.WithDurationAnswer("What is your latency threshold (in milliseconds)?")
		sl.Criteria = &manifest.LatencyCriteria{Threshold: threshold}
	}

	sl.Objective = question.WithFloat64Answer("What is your target for this SLO (in %)?", 0, 100)

	do := question.WithBoolAnswer("Do you want to add an SLI to measure this SLO?")
	if do {
		providers := []string{}
		for key := range providersMap {
			providers = append(providers, key)
		}
		sort.Strings(providers) // sorts slice in-place

		for do {
			providerFullName := question.WithSingleChoiceAnswer("On which cloud provider?", providers...)
			provider := providersMap[providerFullName]
			id := getResourceIDForProvider(provider)

			sl.Indicators = append(sl.Indicators, manifest.ServiceLevelIndicator{
				Provider: provider,
				ID:       id,
			})

			do = question.WithBoolAnswer("Do you want to add another SLI?")
		}
	}
	sl.Name = question.WithStringAnswer("What is the name of this SLO?")
	s.ServiceLevels = append(s.ServiceLevels, &sl)
	fmt.Println(color.Green(fmt.Sprintf("SLO '%s' added to Service '%s'", sl.Name, s.Name)))

	if question.WithBoolAnswer("Do you want to add another SLO?") {
		declareSLOForService(s)
	}
}

func getResourceIDForProvider(provider string) string {
	switch provider {
	case "aws":
		return question.WithStringAnswer("What is the ARN of the resource?")
	case "gcp":
		{
			projectID := question.WithStringAnswer("What is the GCP project ID?")
			resourceType := question.WithSingleChoiceAnswer("What is the 'type' of the resource?", googleResourceTypes...)
			sanitizedResourceType := sanitizeString(resourceType)
			resourceName := question.WithStringAnswer("What is the name of resource?")
			return fmt.Sprintf("%s/%s/%s", projectID, sanitizedResourceType, resourceName)
		}
	default:
		return question.WithStringAnswer("What is the ID of the resource? This could be the AWS ARN, azure resource ID, etc.")
	}
}

func sanitizeString(s string) string {
	return strings.ToLower(strings.ReplaceAll(s, " ", "-"))
}
