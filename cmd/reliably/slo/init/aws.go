package init

import (
	"context"
	"fmt"
	"log"
	"sort"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/apigatewayv2"
	elb "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/reliablyhq/cli/core/cli/question"
)

var awsOptions = []question.AskOpt{question.Subquestion}

func buildAWSArn() string {
	resourceArn := question.WithStringAnswer("Paste an AWS ARN, or type \"i\" for interactive mode.", awsOptions)
	if resourceArn == "i" {
		resolver := endpoints.DefaultResolver()
		partitions := resolver.(endpoints.EnumPartitions).Partitions()

		if len(partitions) == 0 {
			fmt.Println(iconWarn, "Reliably couldn't query AWS. Please try again or use normal mode.")
			return ""
		}

		var partitionsIDs = []string{}
		for _, p := range partitions {
			partitionsIDs = append(partitionsIDs, p.ID())
		}
		partitionID := question.WithSingleChoiceAnswer("Select an AWS partition.", awsOptions, partitionsIDs...)

		var partition endpoints.Partition

		for _, p := range partitions {
			if p.ID() == partitionID {
				partition = p
			}
		}

		if len(partition.Regions()) == 0 {
			if len(partitions) == 0 {
				fmt.Println(iconWarn, "Reliably couldn't query AWS. Please try again or use normal mode.")
				return ""
			}
		}
		var regionsIDs = []string{}
		for id := range partition.Regions() {
			regionsIDs = append(regionsIDs, id)
		}
		sort.Strings(regionsIDs)
		regionID := question.WithSingleChoiceAnswer("Select an AWS region.", awsOptions, regionsIDs...)

		cfgIAM, err := config.LoadDefaultConfig(context.TODO())
		if err != nil {
			fmt.Println(iconWarn, "Reliably encountered a problem. Please try again or use normal mode.")
			return ""
		}
		clientIAM := iam.NewFromConfig(
			cfgIAM,
			func(opt *iam.Options) {
				opt.Region = regionID
			},
		)
		iamParams := &iam.GetUserInput{}
		user, err := clientIAM.GetUser(context.TODO(), iamParams)
		if err != nil {
			fmt.Println(iconWarn, "Reliably couldn't authenticate you with AWS. Make sure you are logged in to AWS.")
			return ""
		}
		userArnStr := aws.ToString(user.User.Arn)
		userArn, err := arn.Parse(userArnStr)
		if err != nil {
			fmt.Println(iconWarn, "Reliably encountered a problem. Please try again or use normal mode.")
			return ""
		}
		accountID := userArn.AccountID

		awsServices := []string{}
		for key := range awsServicesMap {
			awsServices = append(awsServices, key)
		}
		sort.Strings(awsServices) // sorts slice in-place
		serviceFullName := question.WithSingleChoiceAnswer("Select an AWS service.", awsOptions, awsServices...)
		service := awsServicesMap[serviceFullName]

		serviceID := selectAWSService(service, regionID)

		if serviceID == "" {
			return ""
		} else {
			resourceArn = "arn:" + partitionID + ":" + service + ":" + regionID + ":" + accountID + ":" + serviceID
		}
	}
	return resourceArn
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
			fmt.Println(iconWarn, "Reliably encountered a problem. Please try again or use normal mode.")
			return ""
		}

		if len(output.Items) == 0 {
			fmt.Println(iconWarn, "Reliably couldn't find any available resources.")
			return ""
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
		apiNiceName := question.WithSingleChoiceAnswer("Select a Resource.", awsOptions, agwApis...)
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
			fmt.Println(iconWarn, "Reliably encountered a problem. Please try again or use normal mode.")
			return ""
		}
		if len(output.LoadBalancers) == 0 {
			fmt.Println(iconWarn, "Reliably couldn't find any available resources.")
			return ""
		}
		var elbs = []string{}
		elbsMap := make(map[string]string)
		for _, lb := range output.LoadBalancers {
			elbArn := aws.ToString(lb.LoadBalancerArn)
			name := aws.ToString(lb.LoadBalancerName)
			parsedArn, err := arn.Parse(elbArn)
			if err != nil {
				fmt.Println(iconWarn, "Reliably encountered a problem. Please try again or use normal mode.")
				return ""
			}
			elbID := parsedArn.Resource
			elbSlice := strings.Split(elbID, "/")
			elb := elbSlice[len(elbSlice)-1]
			niceName := name + " (" + elb + ")"
			elbs = append(elbs, niceName)
			elbsMap[niceName] = elbID
		}
		elbNiceName := question.WithSingleChoiceAnswer("Select a Resource.", awsOptions, elbs...)
		elbID := elbsMap[elbNiceName]
		return elbID
	default:
		fmt.Println(iconWarn, "Reliably encountered a problem. Please try again or use normal mode.")
		return ""
	}
}
