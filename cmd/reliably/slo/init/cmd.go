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
	elb "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/reliablyhq/cli/core/cli/question"
	"github.com/reliablyhq/cli/core/color"
	"github.com/reliablyhq/cli/core/manifest"
	"github.com/spf13/cobra"

	// compute "google.golang.org/api/compute/v1"
	crm "google.golang.org/api/cloudresourcemanager/v3"
	compute "google.golang.org/api/compute/v1"
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
		"API Getaway":           "apigateway",
		"Elastic Load Balancer": "elasticloadbalancing",
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

			serviceID := selectAWSService(service, regionID)

			resourceArn = "arn:" + partitionID + ":" + service + ":" + regionID + ":" + accountID + ":" + serviceID
		}
		return resourceArn
	case "gcp":
		ctx := context.Background()
		crmService, err := crm.NewService(ctx)
		orgsService := crm.NewOrganizationsService(crmService)
		orgs, err := orgsService.Search().Context(ctx).Do()
		if err != nil {
			// TODO
			// Example error
			// 2021/04/20 12:47:36 googleapi: Error 403: The caller does not have permission, forbidden
			// exit status 1
			// log.Fatal(err)
		}
		var orgsList = []string{}
		orgsMap := make(map[string]string)
		for _, o := range orgs.Organizations {
			displayName := o.DisplayName
			id := o.Name
			orgsList = append(orgsList, displayName)
			orgsMap[displayName] = id
		}
		// TODO Handle empty list case
		orgDisplayName := question.WithSingleChoiceAnswer("Select an Organization.", orgsList...)
		orgID := orgsMap[orgDisplayName]

		projectsService := crm.NewProjectsService(crmService)
		projects, err := projectsService.List().Context(ctx).Parent(orgID).Do()
		if err != nil {
			// Example error if CLoud Resource Manager API has not been used in project or is disabled
			// -----
			// #: &googleapi.Error{Code:403, Message:"Cloud Resource Manager API has not been used in project 473344846455 before or it is disabled. Enable it by visiting https://console.developers.google.com/apis/api/cloudresourcemanager.googleapis.com/overview?project=473344846455 then retry. If you enabled this API recently, wait a few minutes for the action to propagate to our systems and retry.", Details:[]interface {}(nil), Body:"{\n  \"error\": {\n    \"code\": 403,\n    \"message\": \"Cloud Resource Manager API has not been used in project 473344846455 before or it is disabled. Enable it by visiting https://console.developers.google.com/apis/api/cloudresourcemanager.googleapis.com/overview?project=473344846455 then retry. If you enabled this API recently, wait a few minutes for the action to propagate to our systems and retry.\",\n    \"errors\": [\n      {\n        \"message\": \"Cloud Resource Manager API has not been used in project 473344846455 before or it is disabled. Enable it by visiting https://console.developers.google.com/apis/api/cloudresourcemanager.googleapis.com/overview?project=473344846455 then retry. If you enabled this API recently, wait a few minutes for the action to propagate to our systems and retry.\",\n        \"domain\": \"usageLimits\",\n        \"reason\": \"accessNotConfigured\",\n        \"extendedHelp\": \"https://console.developers.google.com\"\n      }\n    ],\n    \"status\": \"PERMISSION_DENIED\"\n  }\n}\n", Header:http.Header(nil), Errors:[]googleapi.ErrorItem{googleapi.ErrorItem{Reason:"accessNotConfigured", Message:"Cloud Resource Manager API has not been used in project 473344846455 before or it is disabled. Enable it by visiting https://console.developers.google.com/apis/api/cloudresourcemanager.googleapis.com/overview?project=473344846455 then retry. If you enabled this API recently, wait a few minutes for the action to propagate to our systems and retry."}}}
			// -----
			// +: googleapi: Error 403: Cloud Resource Manager API has not been used in project 473344846455 before or it is disabled. Enable it by visiting https://console.developers.google.com/apis/api/cloudresourcemanager.googleapis.com/overview?project=473344846455 then retry. If you enabled this API recently, wait a few minutes for the action to propagate to our systems and retry., accessNotConfigured
			// -----
			// *googleapi.Error
			// log.Fatal(err)
		}
		var projectsList = []string{}
		projectsMap := make(map[string]string)
		for _, p := range projects.Projects {
			displayName := p.DisplayName
			id := p.ProjectId
			fullName := displayName + " (" + id + ")"
			projectsList = append(projectsList, fullName)
			projectsMap[fullName] = id
		}
		// TODO Handle empty list case
		projectFullName := question.WithSingleChoiceAnswer("Select an Project.", projectsList...)
		projectID := projectsMap[projectFullName]

		resourceType := question.WithSingleChoiceAnswer("What is the 'type' of the resource?", googleResourceTypes...)
		sanitizedResourceType := sanitizeResourceType(resourceType)

		lbctx := context.Background()
		computeService, err := compute.NewService(lbctx)
		lbsService := compute.NewUrlMapsService(computeService)
		lbs, err := lbsService.List(projectID).Context(ctx).Do()
		if err != nil {
			// TODO
			// log.Fatal(err)
		}
		var lbsList = []string{}
		for _, lb := range lbs.Items {
			name := lb.Name
			lbsList = append(lbsList, name)
		}
		resourceName := question.WithSingleChoiceAnswer("Select a resource.", lbsList...)

		resourceID := fmt.Sprintf("%s/%s/%s", projectID, sanitizedResourceType, resourceName)

		fmt.Println(resourceID)
		return resourceID

	default:
		return question.WithStringAnswer("What is the ID of the resource? This could be the AWS ARN, azure resource ID, etc.")
	}
}

func selectAWSService(serviceType string, region string) string {
	config, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatal(err)
	}

	switch serviceType {
	case "apigateway":
		agwClient := apigatewayv2.NewFromConfig(config)
		output, err := agwClient.GetApis(
			context.TODO(),
			&apigatewayv2.GetApisInput{
				MaxResults: aws.String("50"),
				// NextToken: aws.String("2"),
			},
			func(opt *apigatewayv2.Options) {
				opt.Region = region
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

		return "/apis/" + apiID
	case "elasticloadbalancing":
		elbClient := elb.NewFromConfig(config)
		output, err := elbClient.DescribeLoadBalancers(
			context.TODO(),
			&elb.DescribeLoadBalancersInput{
				// PageSize: aws.String("50"),
				// NextToken:,
			},
			func(opt *elb.Options) {
				opt.Region = region
			},
		)
		if err != nil {
			log.Fatal(err)
		}
		var elbs = []string{}
		elbsMap := make(map[string]string)
		for _, lb := range output.LoadBalancers {
			elbArn := aws.ToString(lb.LoadBalancerArn)
			name := aws.ToString(lb.LoadBalancerName)
			parsedArn, err := arn.Parse(elbArn)
			if err != nil {
				log.Fatal(err)
			}
			elbID := parsedArn.Resource
			elbSlice := strings.Split(elbID, "/")
			elb := elbSlice[len(elbSlice)-1]
			niceName := name + " (" + elb + ")"
			elbs = append(elbs, niceName)
			elbsMap[niceName] = elbID
		}
		elbNiceName := question.WithSingleChoiceAnswer("Select a Resource.", elbs...)
		elbID := elbsMap[elbNiceName]
		return elbID
	default:
		return "something went wrong"
	}
}

func sanitizeResourceType(s string) string {
	return strings.ToLower(strings.ReplaceAll(s, " ", "-"))
}
