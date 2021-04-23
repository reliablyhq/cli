package init

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/reliablyhq/cli/api"
	"github.com/reliablyhq/cli/core"
	"github.com/reliablyhq/cli/core/cli/question"
	"github.com/reliablyhq/cli/core/color"
	"github.com/reliablyhq/cli/core/manifest"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

var (
	manifestPath        string
	service             string
	supportedExtensions = []string{".yaml", ".json"}
	googleResourceTypes = []string{"Google Cloud Load Balancers"}
	awsPartitionsIDs    = []string{
		"aws",
		"aws-cn",
		"aws-us-gov",
	}
	awsServicesMap = map[string]string{
		"API Gateway":           "apigateway",
		"Elastic Load Balancer": "elasticloadbalancing",
	}
	providersMap = map[string]string{
		"Amazon Web Services":   "aws",
		"Google Cloud Platform": "gcp",
	}
)

const iconWarn = "⚠️ "

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

	cmd.Flags().StringVarP(&manifestPath, "path", "p", "", "the location of the manifest file")

	return &cmd
}

func runE(_ *cobra.Command, args []string) error {

	hostname := core.Hostname()
	client := api.NewClientFromHTTP(api.AuthHTTPClient(hostname))

	// Sends the context to Reliably prior executing the command
	orgID, err := api.CurrentUserOrganizationID(client, hostname)
	if err != nil {
		return err
	}
	log.Debug("fetching internal service manifest")
	m, err := api.PullServiceManifest(orgID, service)
	if err != nil {
		return err
	}

	if m == nil {
		log.Debug("no services detected")
		m = &manifest.Manifest{}
	}

	if doesManifestExist(orgID, service) {
		if !question.WithBoolAnswer(fmt.Sprintf("Existing manifest detected (%s); Do you want to overwrite it?", manifestPath), question.WithNoAsDefault) {
			return nil
		}
	}

	populateManifestInteractively(m)

	// validate
	if err := m.Validate(); err != nil {
		return err
	}

	f, err := os.Create(manifestPath)
	if err != nil {
		return err
	}
	defer f.Close()

	if err := yaml.NewEncoder(f).Encode(m); err != nil {
		return err
	}

	// push manifestto backend
	if err := api.PushServiceManifest(orgID, service, m); err != nil {
		return fmt.Errorf("an error occurred while push manifest to reliably: %s", err)
	}

	return nil
}

func populateManifestInteractively(m *manifest.Manifest) {

	var s manifest.Service
	for {
		s.Name = question.WithStringAnswer("What is the name of the service you want to declare SLOs for?")
		if err := validateServiceName(m, s.Name); err != nil {
			fmt.Println(err)
			continue
		}
		break
	}

	declareSLOForService(&s)

	m.Services = append(m.Services, &s)
	fmt.Println(color.Green(fmt.Sprintf("Service '%s' added", s.Name)))

	if question.WithBoolAnswer("Do you want to add another Service?", question.WithNoAsDefault) {
		populateManifestInteractively(m)
	}

}

func declareSLOForService(s *manifest.Service) {
	var sl manifest.ServiceLevel

	slType := question.WithSingleChoiceAnswer("What type of SLO do you want to declare?", "Availability", "Latency")
	sl.Type = sanitizeString(slType)

	sl.Objective = question.WithFloat64Answer("What is your target for this SLO (in %)?", 0, 100)

	if sl.Type == "latency" {
		threshold := question.WithDurationAnswer("What is your latency threshold (in milliseconds)?")
		sl.Criteria = &manifest.LatencyCriteria{Threshold: threshold}
	}

	do := question.WithBoolAnswer("Do you want to add a resource for measuring your SLI?", question.WithYesAsDefault)
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

			if id != "" { // We're returning empty strings when something fails...
				sl.Indicators = append(sl.Indicators, manifest.ServiceLevelIndicator{
					Provider: provider,
					ID:       id,
				})
			}

			do = question.WithBoolAnswer("Do you want to add another resource for measuring your SLI?", question.WithNoAsDefault)
		}
	}

	sl.Name = question.WithStringAnswer("What is the name of this SLO?")

	s.ServiceLevels = append(s.ServiceLevels, &sl)
	fmt.Println(color.Green(fmt.Sprintf("SLO '%s' added to Service '%s'", sl.Name, s.Name)))

	if question.WithBoolAnswer("Do you want to add another SLO?", question.WithNoAsDefault) {
		declareSLOForService(s)
	}
}

func getResourceIDForProvider(provider string) string {
	switch provider {
	case "aws":
		return buildAWSArn()
	case "gcp":
		return buildGCPResourceID()
	default:
		return question.WithStringAnswer("What is the ID of the resource? This could be the AWS ARN, azure resource ID, etc.")
	}
}

func sanitizeString(s string) string {
	return strings.ToLower(strings.ReplaceAll(s, " ", "-"))
}

// -- validation functions

func validateFilePath() error {
	for _, ext := range supportedExtensions {
		if strings.HasSuffix(manifestPath, ext) {
			return nil
		}
	}

	return fmt.Errorf("manifest file must have one of the these extensions: %v", supportedExtensions)
}

func validateServiceName(m *manifest.Manifest, name string) error {
	for _, s := range m.Services {
		if name == s.Name {
			return fmt.Errorf("service name [%s] already exists, please enter another", color.Red(name))
		}
	}
	return nil
}

func doesManifestExist(org, service string) bool {
	if found, _ := api.ServiceExists(org, service); found {
		return true
	}

	if manifestPath != "" {
		if _, err := os.Stat(manifestPath); err == nil {
			return true
		}
	}

	return false
}
