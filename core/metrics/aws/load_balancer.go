package aws

import (
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
	log "github.com/sirupsen/logrus"
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

func (elb *ElasticLoadBalancer) GetLatencyAboveThresholdPerMin(
	arn arn.ARN, threshold float64, from, to time.Time) (*cloudwatch.GetMetricDataInput, error) {

	var params *cloudwatch.GetMetricDataInput

	ns := elb.Namespace()
	dim, err := elb.Dimension(arn)
	if err != nil {
		return nil, err
	}

	period := int32(to.Sub(from).Seconds())

	///
	//from, _ = time.Parse(time.RFC3339, "2021-04-12T15:00:00Z")
	//to, _ = time.Parse(time.RFC3339, "2021-04-13T08:00:00Z")
	//period = int32(to.Sub(from).Seconds())

	fmt.Println("from", from, " - to ", to, " => ", period)
	fmt.Println("how many minutes in the period ? ", period/60)
	///

	expression := fmt.Sprintf("latency_p99_per_min > %f", threshold)
	log.Debugf("Latency expression=%s\n", expression)

	params = &cloudwatch.GetMetricDataInput{
		StartTime: &from,
		EndTime:   &to,
		MetricDataQueries: []types.MetricDataQuery{

			{
				Id:         aws.String("latency_above_threshold_per_min"),
				Expression: aws.String(expression),
			},

			{
				Id: aws.String("latency_p99_per_min"),
				MetricStat: &types.MetricStat{
					Metric: &types.Metric{
						Namespace:  aws.String(ns),
						MetricName: aws.String("TargetResponseTime"),
						Dimensions: []types.Dimension{dim},
					},
					Period: aws.Int32(60),
					Stat:   aws.String("p99"),
				},
				ReturnData: aws.Bool(false),
			},
		},
	}

	return params, nil

}
