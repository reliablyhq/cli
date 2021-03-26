package metrics

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
	log "github.com/sirupsen/logrus"
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
	log.Debugf("%#v", res)

	params, err := res.GetMetricDataInput(from, to)
	if err != nil {
		return -1, err
	}

	log.Debugf("Retrieve metrics From %s To %s", from, to)
	data, err := cw.client.GetMetricData(context.TODO(), params)
	if err != nil {
		return -1, err
	}

	var errorRatePercent float64 = -1
	for _, r := range data.MetricDataResults {
		if string(*r.Id) == "error_rate_percent" {
			if len(r.Values) > 0 {
				errorRatePercent = r.Values[0]
			}
			break
		}
	}
	if errorRatePercent == -1 {
		return errorRatePercent, errors.New("No error rate percent value retrieved from cloud watch")
	}

	log.Debugf("error rate is %v%%\n", errorRatePercent)
	return errorRatePercent, nil
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
func (r *AwsResource) IsSupportedService() bool {
	switch r.arn.Service {
	case "apigateway":
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

func (r *AwsResource) MetricNamespace() string {

	ns := ""

	switch r.arn.Service {
	case "apigateway":
		ns = "AWS/ApiGateway"
	case "cloudfront":
		ns = "AWS/CloudFront"
	case "elasticloadbalancing":
		ns = "AWS/ApplicationELB"
	}

	return ns
}

// GetMetricDataInput retuns the MetricDataInput struct for querying
// with cloud watch API. It must ONLY return a single value for the
// required 'error_rate_percent' metric ID
// This function handles data input depending on different targeted AWS Service
func (r *AwsResource) GetMetricDataInput(from, to time.Time) (*cloudwatch.GetMetricDataInput, error) {

	var params *cloudwatch.GetMetricDataInput

	period := int32(to.Sub(from).Seconds())
	ns := r.MetricNamespace()

	switch r.arn.Service {
	case "apigateway":

		parts := strings.Split(r.arn.Resource, "/")
		apiID := parts[len(parts)-1]

		params = &cloudwatch.GetMetricDataInput{
			StartTime: &from,
			EndTime:   &to,
			MetricDataQueries: []types.MetricDataQuery{

				{
					Id:         aws.String("error_rate_percent"),
					Expression: aws.String("error_rate * 100"),
				},

				{
					Id:         aws.String("error_rate"),
					Expression: aws.String("errors / requests"),
					ReturnData: aws.Bool(false),
				},

				{
					Id:         aws.String("errors"),
					Expression: aws.String("SUM([http_5xx_error_count, http_4xx_error_count])"),
					ReturnData: aws.Bool(false),
				},

				{
					Id: aws.String("requests"),
					MetricStat: &types.MetricStat{
						Metric: &types.Metric{
							Namespace:  aws.String(ns),
							MetricName: aws.String("Count"),
							Dimensions: []types.Dimension{
								{
									Name:  aws.String("ApiId"),
									Value: aws.String(apiID),
								},
							},
						},
						Period: aws.Int32(period),
						Stat:   aws.String("SampleCount"),
					},
					ReturnData: aws.Bool(false),
				},

				{
					Id: aws.String("http_5xx_error_count"),
					MetricStat: &types.MetricStat{
						Metric: &types.Metric{
							Namespace:  aws.String(ns),
							MetricName: aws.String("5xx"),
							Dimensions: []types.Dimension{
								{
									Name:  aws.String("ApiId"),
									Value: aws.String(apiID),
								},
							},
						},
						Period: aws.Int32(period),
						Stat:   aws.String("Sum"),
					},
					ReturnData: aws.Bool(false),
				},

				{
					Id: aws.String("http_4xx_error_count"),
					MetricStat: &types.MetricStat{
						Metric: &types.Metric{
							Namespace:  aws.String(ns),
							MetricName: aws.String("4xx"),
							Dimensions: []types.Dimension{
								{
									Name:  aws.String("ApiId"),
									Value: aws.String(apiID),
								},
							},
						},
						Period: aws.Int32(period),
						Stat:   aws.String("Sum"),
					},
					ReturnData: aws.Bool(false),
				},
			},
		}

	default:
		return nil, fmt.Errorf("Unable to construct Metric Query for AWS Service '%s'", r.arn.Service)
	}

	return params, nil

}
