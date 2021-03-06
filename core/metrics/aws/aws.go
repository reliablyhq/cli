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

	"github.com/reliablyhq/cli/core/entities"
	"github.com/reliablyhq/cli/utils"
)

// SelectorKeys - a list of all
// supported AWS specific selector keys
var SelectorKeys = []string{
	"aws_arn",
}

type AwsCloudWatch struct{}

type AwsMetricsProvider interface {
	Namespace() string
	Dimension(arn.ARN) (types.Dimension, error)
	GetErrorRateMetricDataInput(arn.ARN, time.Time, time.Time) (*cloudwatch.GetMetricDataInput, error)
	GetLatencyMetricDataInput(arn.ARN, time.Time, time.Time) (*cloudwatch.GetMetricDataInput, error)
	GetLatencyAboveThresholdPerMin(arn.ARN, time.Time, time.Time, float64) (*cloudwatch.GetMetricDataInput, error)
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
	// Credentials to AWS go SDK can be setup as described in the official doc:
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

// Nothing to do for AWS...
// We don't want to close anything as providers are cached.
func (cw *AwsCloudWatch) Close() error {
	return nil
}

func (cw *AwsCloudWatch) GetLatencyAboveThresholdPercentage(resourceID string, from, to time.Time, threshold int) (float64, error) {
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

	thresholdf64 := float64(threshold)
	switch res.arn.Service {
	case "elasticloadbalancing":
		// TargetResponseTime is returned as seconds not ms
		thresholdf64 = thresholdf64 / 1000
	}
	params, err := provider.GetLatencyAboveThresholdPerMin(res.arn, from, to, thresholdf64)
	if err != nil {
		return -1, err
	}

	log.Debugf("Retrieve latency above %vms threshold metrics From %s To %s", threshold, from, to)
	data, err := client.GetMetricData(ctx, params)
	if err != nil {
		return -1, err
	}

	var latencyAboveThreshold float64 = -1
	for _, r := range data.MetricDataResults {
		if string(*r.Id) == "latency_above_threshold_per_min" {
			if len(r.Values) > 0 {
				latencyAboveThreshold = utils.SumFloat64(r.Values) / float64(len(r.Values)) * 100.0
			}
			break
		}
	}
	if latencyAboveThreshold == -1 {
		return latencyAboveThreshold, errors.New("No latency value retrieved from cloud watch")
	}

	log.Debugf("Latency above threshold %vms is %.2f%%\n", threshold, latencyAboveThreshold)
	return latencyAboveThreshold, nil
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

func (cw *AwsCloudWatch) GetAvailabilityPercentage(resourceID string, from, to time.Time) (float64, error) {
	errorRate, err := cw.GetErrorPercentageMetricForResource(resourceID, from, to)
	if err != nil {
		return -1, err
	}

	availability := 100.0 - errorRate
	log.Debugf("Availability is %.2f%%\n", availability)
	return availability, nil
}

// ResourceFromSelector - identifies the resource ID given a selector.
func (cw *AwsCloudWatch) ResourceFromSelector(s entities.Selector) string {
	if v, ok := s["aws_arn"]; ok {
		return v
	}

	return ""
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
		provider = &ApiGateway{}
	case "elasticloadbalancing":
		provider = &ElasticLoadBalancer{}
	default:
		err = fmt.Errorf("Service %s is not currently supported", r.arn.Service)
	}

	return
}

func (cw *AwsCloudWatch) CanHandleSelector(labels entities.Selector) bool {
	resourceID := cw.ResourceFromSelector(labels)
	if resourceID == "" {
		return false
	}

	return true
}

func (cw *AwsCloudWatch) ComputeObjective(o *entities.Objective, from time.Time, to time.Time) (*entities.Indicator, error) {

	i := entities.NewIndicatorForObjective(o, from, to)
	var err error

	resourceID := cw.ResourceFromSelector(o.Spec.IndicatorSelector)
	if resourceID == "" {
		return nil, fmt.Errorf("unable to identify provider and resource id for objective: %v",
			o.Spec.IndicatorSelector)
	}

	switch o.Spec.IndicatorSelector["category"] {
	case "latency":
		target, ok := o.Spec.IndicatorSelector["latency_target"]
		if !ok {
			return nil, errors.New("latency_target not defined in Objective spec")
		}

		thres, err := time.ParseDuration(target)
		if err != nil {
			return nil, err
		}

		i.Spec.Percent, err = cw.GetLatencyAboveThresholdPercentage(
			resourceID, from, to, int(thres.Milliseconds()))
		if err != nil {
			return nil, err
		}

	case "availability":
		i.Spec.Percent, err = cw.GetAvailabilityPercentage(
			resourceID, from, to)
		if err != nil {
			return nil, err
		}

	default:
		return nil, fmt.Errorf("unsupported indicator category: %s",
			o.Spec.IndicatorSelector["category"])
	}

	return i, nil
}
