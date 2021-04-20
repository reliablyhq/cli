package aws

import (
	"reflect"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	//"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	"github.com/stretchr/testify/assert"
)

func TestELBNamespace(t *testing.T) {
	elb := &ElasticLoadBalancer{}
	assert.NotEqual(t, "", elb.Namespace(), "Namespace should not be empty string")
}

func TestELBDimension(t *testing.T) {
	arnStr := "arn:aws:elasticloadbalancing:eu-west-1:123456789:loadbalancer/app/dummy/az12er34ty56"
	arn, _ := arn.Parse(arnStr)

	elb := &ElasticLoadBalancer{}
	dim, err := elb.Dimension(arn)

	assert.NoError(t, err)
	assert.NotEqual(t, nil, dim, "Dimension has not been returned")
	assert.Equal(t, "app/dummy/az12er34ty56", aws.ToString(dim.Value), "Unexpected ELB ID from dimension")
}

func TestELBGetMetricQueryInputs(t *testing.T) {

	to := time.Now()
	from := to.Add(-oneDay)
	arn, _ := arn.Parse(arnELB)

	elb := &ElasticLoadBalancer{}

	type args []interface{}
	tests := []struct {
		name string
		f    reflect.Value
		args args
	}{
		{
			name: "error rate",
			f:    reflect.ValueOf(elb.GetErrorRateMetricDataInput),
			args: args{
				arn, from, to,
			},
		},
		{
			name: "p99 latency",
			f:    reflect.ValueOf(elb.GetLatencyMetricDataInput),
			args: args{
				arn, from, to,
			},
		},
		{
			name: "latency above threshold",
			f:    reflect.ValueOf(elb.GetLatencyAboveThresholdPerMin),
			args: args{
				arn, from, to, 250.0,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var args []reflect.Value
			for _, x := range tt.args {
				args = append(args, reflect.ValueOf(x))
			}

			values := tt.f.Call(args)

			m, err := values[0], values[1]
			assert.Equal(t, nil, err.Interface(), "Unexpected error")
			assert.NotEqual(t, nil, m, "Metric Query should not be nil")
		})
	}
}
