package aws

import (
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
)

type ElasticLoadBalancer struct{}

func (elb *ElasticLoadBalancer) Namespace() string {
	return "AWS/ApplicationELB"
}

func (elb *ElasticLoadBalancer) Dimension(arn arn.ARN) (types.Dimension, error) {

	parts := strings.Split(arn.Resource, "/")
	lbID := strings.Join(parts[1:], "/")

	dim := types.Dimension{
		Name:  aws.String("LoadBalancer"),
		Value: aws.String(lbID),
	}

	return dim, nil
}

func (elb *ElasticLoadBalancer) GetErrorRateMetricDataInput(arn arn.ARN, from, to time.Time) (*cloudwatch.GetMetricDataInput, error) {

	var params *cloudwatch.GetMetricDataInput

	period := int32(to.Sub(from).Seconds())

	ns := elb.Namespace()
	dim, err := elb.Dimension(arn)
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
				Expression: aws.String("SUM([http_5xx_error_count, elb_5xx_error_count])"),
				ReturnData: aws.Bool(false),
			},

			{
				Id: aws.String("requests"),
				MetricStat: &types.MetricStat{
					Metric: &types.Metric{
						Namespace:  aws.String(ns),
						MetricName: aws.String("RequestCount"),
						Dimensions: []types.Dimension{dim},
					},
					Period: aws.Int32(period),
					Stat:   aws.String("Sum"),
				},
				ReturnData: aws.Bool(false),
			},

			{
				Id: aws.String("http_5xx_error_count"),
				MetricStat: &types.MetricStat{
					Metric: &types.Metric{
						Namespace:  aws.String(ns),
						MetricName: aws.String("HTTPCode_Target_5XX_Count"),
						Dimensions: []types.Dimension{dim},
					},
					Period: aws.Int32(period),
					Stat:   aws.String("Sum"),
				},
				ReturnData: aws.Bool(false),
			},

			{
				Id: aws.String("elb_5xx_error_count"),
				MetricStat: &types.MetricStat{
					Metric: &types.Metric{
						Namespace:  aws.String(ns),
						MetricName: aws.String("HTTPCode_ELB_5XX_Count"),
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

func (elb *ElasticLoadBalancer) GetLatencyMetricDataInput(arn arn.ARN, from, to time.Time) (*cloudwatch.GetMetricDataInput, error) {

	var params *cloudwatch.GetMetricDataInput

	period := int32(to.Sub(from).Seconds())

	ns := elb.Namespace()
	dim, err := elb.Dimension(arn)
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
						MetricName: aws.String("TargetResponseTime"),
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
