package aws

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
	log "github.com/sirupsen/logrus"

	apiGateway "github.com/reliablyhq/cli/core/metrics/aws/api_gateway"
	elb "github.com/reliablyhq/cli/core/metrics/aws/elb"
)

type AwsCloudWatch struct{}

type AwsMetricsProvider interface {
	Namespace() string
	Dimension(arn.ARN) (types.Dimension, error)
	GetErrorRateMetricDataInput(arn.ARN, time.Time, time.Time) (*cloudwatch.GetMetricDataInput, error)
	GetLatencyMetricDataInput(arn.ARN, time.Time, time.Time) (*cloudwatch.GetMetricDataInput, error)
}

type AwsResource struct {
	arn arn.ARN
}

type AwsCloudWatchClient interface {
	GetMetricData(ctx context.Context, params *cloudwatch.GetMetricDataInput,
		optFns ...func(*cloudwatch.Options)) (*cloudwatch.GetMetricDataOutput, error)
}

var (
	ctx               = context.TODO()
	cloudwatchClients = map[string]AwsCloudWatchClient{}
)

// NewAwsCloudWatch is the factory function for AWS cloud watch metric provider
func NewAwsCloudWatch() (cw *AwsCloudWatch, err error) {
	// Credentials to AWS go SDK can be setup as described in the offical doc:
	// https://aws.github.io/aws-sdk-go-v2/docs/configuring-sdk/#specifying-credentials

	return &AwsCloudWatch{}, nil
}

func tryGetClient(region string) (AwsCloudWatchClient, error) {
	var ok bool
	var client AwsCloudWatchClient
	if client, ok = cloudwatchClients[region]; !ok {
		cfg, err := config.LoadDefaultConfig(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to load configuration, %v", err)
		}

		setRegion := func(opts *cloudwatch.Options) {
			if region != "" {
				opts.Region = region
			}
		}

		client = cloudwatch.NewFromConfig(cfg, setRegion)
		cloudwatchClients[region] = client
	}

	return client, nil
}

func (cw *AwsCloudWatch) Get99PercentLatencyMetricForResource(resourceID string, from, to time.Time) (float64, error) {
	res, err := parseResourceID(resourceID)
	if err != nil {
		return -1, err
	}
	log.Debugf("%#v", res)

	client, err := tryGetClient(res.arn.Region)
	if err != nil {
		return -1, err
	}

	provider, err := res.MetricProvider()
	if err != nil {
		return -1, err
	}

	params, err := provider.GetLatencyMetricDataInput(res.arn, from, to)
	if err != nil {
		return -1, err
	}

	log.Debugf("Retrieve latency metrics From %s To %s", from, to)
	data, err := client.GetMetricData(ctx, params)
	if err != nil {
		return -1, err
	}

	var latencyPercentile float64 = -1
	for _, r := range data.MetricDataResults {
		if string(*r.Id) == "latency_percentile" {
			if len(r.Values) > 0 {
				latencyPercentile = r.Values[0]
			}

			for i, ts := range r.Timestamps {
				log.Debugf("> %v : %.3fms", ts, r.Values[i])
			}
			break
		}
	}
	if latencyPercentile == -1 {
		return latencyPercentile, errors.New("No latency value retrieved from cloud watch")
	}

	switch res.arn.Service {
	case "elasticloadbalancing":
		// TargetResponseTime is returned as seconds not ms
		latencyPercentile = latencyPercentile * 1000
	}

	log.Debugf("99 percentile latency is %.3fms\n", latencyPercentile)
	return latencyPercentile, nil
}

func (cw *AwsCloudWatch) GetErrorPercentageMetricForResource(resourceID string, from, to time.Time) (float64, error) {
	res, err := parseResourceID(resourceID)
	if err != nil {
		return -1, err
	}
	log.Debugf("%#v", res)

	client, err := tryGetClient(res.arn.Region)
	if err != nil {
		return -1, err
	}

	provider, err := res.MetricProvider()
	if err != nil {
		return -1, err
	}

	params, err := provider.GetErrorRateMetricDataInput(res.arn, from, to)
	if err != nil {
		return -1, err
	}

	log.Debugf("Retrieve error rate metrics From %s To %s", from, to)
	data, err := client.GetMetricData(ctx, params)
	if err != nil {
		return -1, err
	}

	var errorRatePercent float64 = -1
	for _, r := range data.MetricDataResults {
		if string(*r.Id) == "error_rate_percent" {
			if len(r.Values) > 0 {
				errorRatePercent = r.Values[0]
			}

			for i, ts := range r.Timestamps {
				log.Debugf("> %v : %.2f%%", ts, r.Values[i])
			}
			break
		}
	}
	if errorRatePercent == -1 {
		return errorRatePercent, errors.New("No error rate percent value retrieved from cloud watch")
	}

	log.Debugf("error rate is %.2f%%\n", errorRatePercent)
	return errorRatePercent, nil
}

// IsSupportedService indicates wether the resource is supported
// for metrics retrieval
func (r *AwsResource) IsSupportedService() bool {
	switch r.arn.Service {
	case "apigateway", "elasticloadbalancing":
		return true
	default:
		return false
	}
}

func parseResourceID(resId string) (AwsResource, error) {

	zerovalue := AwsResource{}

	if !arn.IsARN(resId) {
		return zerovalue, fmt.Errorf("Resource ID '%s' is not a valid ARN", resId)
	}

	arn, err := arn.Parse(resId)
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

func (r *AwsResource) MetricProvider() (provider AwsMetricsProvider, err error) {

	switch r.arn.Service {
	case "apigateway":
		provider = &apiGateway.Provider{}
	case "elasticloadbalancing":
		provider = &elb.Provider{}
	default:
		err = fmt.Errorf("No metric provider found for AWS Service '%s'", r.arn.Service)
	}

	return
}
