package metrics

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	//"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	//"github.com/aws/aws-sdk-go-v2/aws/session"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
)

type AwsCloudWatch struct {
	client *cloudwatch.Client
	config aws.Config
}

type AwsResource struct {
	arn arn.ARN
}

// NewAwsCloudWatch is the factory function for AWS cloud watch metric provider
func NewAwsCloudWatch() (cw *AwsCloudWatch, err error) {

	/*
		sess := session.Must(session.NewSessionWithOptions(session.Options{
			SharedConfigState: session.SharedConfigEnable,
		}))
		fmt.Println("aws session ", sess)
	*/

	// Credentials to AWS go SDK can be setup as described in the offical doc:
	// https://aws.github.io/aws-sdk-go-v2/docs/configuring-sdk/#specifying-credentials

	cw = &AwsCloudWatch{}

	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		err = fmt.Errorf("failed to load configuration, %v", err)
		return
	}
	cw.config = cfg

	cw.client = cloudwatch.NewFromConfig(cfg)

	fmt.Println("AWS Cloud Watch -> ", cw)
	fmt.Printf("cloud watch client -> %#v\n", cw.client)
	return
}

func (cw *AwsCloudWatch) Get99PercentLatencyMetricForResource(resourceID string, from, to time.Time) (float64, error) {
	return -1, nil
}

func (cw *AwsCloudWatch) GetErrorPercentageMetricForResource(resourceID string, from, to time.Time) (float64, error) {
	res, err := parseResourceID(resourceID)
	if err != nil {
		return -1, err
	}

	fmt.Println(res.arn.Resource)
	parts := strings.Split(res.arn.Resource, "/")
	apiID := parts[len(parts)-1]
	fmt.Println("resource iD for api ", apiID)

	var (
	//id string = "http_error_count"
	//expression string = "SUM([METRICS(\"4xx\"), METRICS(\"5xx\")])"
	)

	//"AWS/ApiGateway", "4xx", "ApiId", "trj7cyiqib"

	fmt.Println(from, to)

	params := &cloudwatch.GetMetricDataInput{
		StartTime: &from,
		EndTime:   &to,
		MetricDataQueries: []types.MetricDataQuery{
			/*
				{
					Id: aws.String("http_error_count"),
					//Expression: aws.String("SUM([METRICS(\"4xx\"), METRICS(\"5xx\")])"),
					//Expression: &expression,
					MetricStat: &types.MetricStat{
						Metric: &types.Metric{
							Namespace:  aws.String("AWS/ApiGateway"),
							MetricName: aws.String("4xx"),
						},
					},
				},
			*/
			{
				Id:         aws.String("http_error_count"),
				Expression: aws.String("SUM([http_5xx_error_count, http_4xx_error_count])"),
			},
			{
				Id: aws.String("http_5xx_error_count"),
				//Expression: aws.String("SUM([METRICS(\"4xx\"), METRICS(\"5xx\")])"),
				MetricStat: &types.MetricStat{
					Metric: &types.Metric{
						Namespace:  aws.String("AWS/ApiGateway"),
						MetricName: aws.String("5xx"),
						Dimensions: []types.Dimension{
							{
								Name:  aws.String("ApiId"),
								Value: aws.String(apiID),
							},
						},
					},
					Period: aws.Int32(60),
					Stat:   aws.String("Sum"),
				},
				ReturnData: aws.Bool(false),
			},
			{
				Id: aws.String("http_4xx_error_count"),
				//Expression: aws.String("SUM([METRICS(\"4xx\"), METRICS(\"5xx\")])"),
				MetricStat: &types.MetricStat{
					Metric: &types.Metric{
						Namespace:  aws.String("AWS/ApiGateway"),
						MetricName: aws.String("4xx"),
						Dimensions: []types.Dimension{
							{
								Name:  aws.String("ApiId"),
								Value: aws.String(apiID),
							},
						},
					},
					Period: aws.Int32(60),
					Stat:   aws.String("Sum"),
				},
				ReturnData: aws.Bool(false),
			},
		},
	}
	data, err := cw.client.GetMetricData(context.TODO(), params)
	if err != nil {
		return -1, err
	}

	fmt.Printf("metric data: %#v\n", data)
	for _, r := range data.MetricDataResults {

		fmt.Println(string(*r.Id), string(*r.Label))

		for i, ts := range r.Timestamps {

			fmt.Printf("%v -> %v\n", ts, r.Values[i])
		}

	}

	return -1, nil
}

// extractArnFromResourceID returns the ARN subpart of a service resource ID
// ie aws/arn:partition:service:region:account-id:resource-type:resource-id
func extractArnFromResourceID(id string) (arn.ARN, error) {

	var arnStr string
	if strings.HasPrefix(id, "aws/") {
		arnStr = strings.SplitN(id, "/", 2)[1] // ID is aws/arn:aws:...
	} else {
		arnStr = id // ID is directly arn:aws:...
	}

	if arnStr == "" {
		return arn.ARN{}, errors.New("Missing ARN in resource identifier: aws/arn:...")
	}

	if !arn.IsARN(arnStr) {
		return arn.ARN{}, fmt.Errorf("'%s' is not a valid ARN", arnStr)
	}

	return arn.Parse(arnStr)
}

// IsSupportedService indicates wether the resource is supported
// for metrics retrieval
func (aws *AwsResource) IsSupportedService() bool {
	switch aws.arn.Service {
	case "apigateway", "apigateway2":
		return true
	default:
		return false
	}
}

func parseResourceID(resId string) (AwsResource, error) {

	zerovalue := AwsResource{}

	arn, err := extractArnFromResourceID(resId)
	if err != nil {
		return zerovalue, err
	}

	resource := AwsResource{arn: arn}
	if !resource.IsSupportedService() {
		return zerovalue,
			fmt.Errorf("AWS Service '%s' is not supported", resource.arn.Service)
	}

	return resource, nil
}
