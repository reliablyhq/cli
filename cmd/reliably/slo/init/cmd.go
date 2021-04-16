package init

import (
	"context"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"

	// Using this as v2 doesn't have an equivalent
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/apigatewayv2"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/aws/endpoints"
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
	awsPartitionsIDs    = []string{
		"aws",
		"aws-cn",
		"aws-us-gov",
	}
	awsServicesMap = map[string]string{
		"API Getaway": "apigateway",
	}
	providersMap = map[string]string{
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
		if !question.WithBoolAnswer(fmt.Sprintf("Existing manifest detected (%s); Do you want to overwrite it?", manifestPath), question.WithNoAsDefault) {
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

			sl.Indicators = append(sl.Indicators, manifest.ServiceLevelIndicator{
				Provider: provider,
				ID:       id,
			})

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
		resourceArn := question.WithStringAnswer("Paste an AWS ARN, or type \"help\" for step-by-step identification.")
		if resourceArn == "help" {
			resolver := endpoints.DefaultResolver()
			partitions := resolver.(endpoints.EnumPartitions).Partitions()

			var partitionsIDs = []string{}
			for _, p := range partitions {
				partitionsIDs = append(partitionsIDs, p.ID())
			}
			partitionID := question.WithSingleChoiceAnswer("Select an AWS partition.", partitionsIDs...)

			var partition endpoints.Partition

			for _, p := range partitions {
				if p.ID() == partitionID {
					partition = p
				}
			}

			var regionsIDs = []string{}
			for id := range partition.Regions() {
				regionsIDs = append(regionsIDs, id)
			}
			sort.Strings(regionsIDs)
			regionID := question.WithSingleChoiceAnswer("Select an AWS region.", regionsIDs...)

			cfgIAM, err := config.LoadDefaultConfig(context.TODO())
			if err != nil {
				log.Fatal(err)
			}
			clientIAM := iam.NewFromConfig(
				cfgIAM,
				func(opt *iam.Options) {
					opt.Region = regionID
				},
			)
			iamParams := &iam.GetUserInput{}
			user, err := clientIAM.GetUser(context.TODO(), iamParams)
			userArnStr := aws.ToString(user.User.Arn)
			userArn, err := arn.Parse(userArnStr)
			accountID := userArn.AccountID

			awsServices := []string{}
			for key := range awsServicesMap {
				awsServices = append(awsServices, key)
			}
			sort.Strings(awsServices) // sorts slice in-place
			serviceFullName := question.WithSingleChoiceAnswer("Select an AWS service.", awsServices...)
			service := awsServicesMap[serviceFullName]

			agwCfg, err := config.LoadDefaultConfig(context.TODO())
			if err != nil {
				log.Fatal(err)
			}
			agwClient := apigatewayv2.NewFromConfig(agwCfg)
			output, err := agwClient.GetApis(
				context.TODO(),
				&apigatewayv2.GetApisInput{
					MaxResults: aws.String("50"),
					// NextToken: aws.String("2"),
				},
				func(opt *apigatewayv2.Options) {
					opt.Region = regionID
				},
			)
			if err != nil {
				log.Fatal(err)
			}

			var agwApis = []string{}
			agwApisMap := make(map[string]string)
			for _, api := range output.Items {
				name := aws.ToString(api.Name)
				ID := aws.ToString(api.ApiId)
				niceName := name + " (" + ID + ")"
				agwApis = append(agwApis, niceName)
				agwApisMap[niceName] = ID
			}
			apiNiceName := question.WithSingleChoiceAnswer("Select a Resource.", agwApis...)
			apiID := agwApisMap[apiNiceName]

			resourceArn = "arn:" + partitionID + ":" + service + ":" + regionID + ":" + accountID + ":/apis/" + apiID
		}
		return resourceArn
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
