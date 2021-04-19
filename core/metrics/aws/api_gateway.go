package aws

import (
	"errors"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
)

type ApiGateway struct{}

func (agw *ApiGateway) Namespace() string {
	return "AWS/ApiGateway"
}

func (agw *ApiGateway) Dimension(arn arn.ARN) (types.Dimension, error) {

	parts := strings.Split(arn.Resource, "/")
	apiID := parts[len(parts)-1]

	dim := types.Dimension{
		Name:  aws.String("ApiId"),
		Value: aws.String(apiID),
	}

	return dim, nil
}

func (agw *ApiGateway) GetErrorRateMetricDataInput(arn arn.ARN, from, to time.Time) (*cloudwatch.GetMetricDataInput, error) {

	var params *cloudwatch.GetMetricDataInput

	period := int32(to.Sub(from).Seconds())

	ns := agw.Namespace()
	dim, err := agw.Dimension(arn)
	if err != nil {
		return nil, err
	}

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
				Expression: aws.String("SUM([http_5xx_error_count])"),
				ReturnData: aws.Bool(false),
			},

			{
				Id: aws.String("requests"),
				MetricStat: &types.MetricStat{
					Metric: &types.Metric{
						Namespace:  aws.String(ns),
						MetricName: aws.String("Count"),
						Dimensions: []types.Dimension{dim},
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
						Dimensions: []types.Dimension{dim},
					},
					Period: aws.Int32(period),
					Stat:   aws.String("Sum"),
				},
				ReturnData: aws.Bool(false),
			},
		},
	}

	return params, nil
}

func (agw *ApiGateway) GetLatencyMetricDataInput(arn arn.ARN, from, to time.Time) (*cloudwatch.GetMetricDataInput, error) {

	var params *cloudwatch.GetMetricDataInput

	period := int32(to.Sub(from).Seconds())

	ns := agw.Namespace()
	dim, err := agw.Dimension(arn)
	if err != nil {
		return nil, err
	}

	params = &cloudwatch.GetMetricDataInput{
		StartTime: &from,
		EndTime:   &to,
		MetricDataQueries: []types.MetricDataQuery{

			{
				Id: aws.String("latency_percentile"),
				MetricStat: &types.MetricStat{
					Metric: &types.Metric{
						Namespace:  aws.String(ns),
						MetricName: aws.String("Latency"),
						Dimensions: []types.Dimension{dim},
					},
					Period: aws.Int32(period),
					Stat:   aws.String("p99"),
				},
			},
		},
	}

	return params, nil
}

func (agw *ApiGateway) GetLatencyAboveThresholdPerMin(
	arn arn.ARN, from, to time.Time, threshold float64) (*cloudwatch.GetMetricDataInput, error) {

	return nil, errors.New("Not implemented yet")
}
