package aws

import (
	"reflect"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/stretchr/testify/assert"
)

func TestApiGWNamespace(t *testing.T) {
	apigw := &ApiGateway{}
	assert.NotEqual(t, "", apigw.Namespace(), "Namespace should not be empty string")
}

func TestApiGWDimension(t *testing.T) {
	arnStr := "arn:aws:apigateway:eu-west-1:123456789:/apis/az12er34ty56"
	arn, _ := arn.Parse(arnStr)

	apigw := &ApiGateway{}
	dim, err := apigw.Dimension(arn)

	assert.NoError(t, err)
	assert.NotEqual(t, nil, dim, "Dimension has not been returned")
	assert.Equal(t, "az12er34ty56", aws.ToString(dim.Value), "Unexpected ApiGateway ID from dimension")
}

func TestApiGWGetMetricQueryInputs(t *testing.T) {

	to := time.Now()
	from := to.Add(-oneDay)
	arn, _ := arn.Parse(arnELB)

	apigw := &ApiGateway{}

	type args []interface{}
	tests := []struct {
		name string
		f    reflect.Value
		args args
	}{
		{
			name: "error rate",
			f:    reflect.ValueOf(apigw.GetErrorRateMetricDataInput),
			args: args{
				arn, from, to,
			},
		},
		{
			name: "p99 latency",
			f:    reflect.ValueOf(apigw.GetLatencyMetricDataInput),
			args: args{
				arn, from, to,
			},
		},
		{
			name: "latency above threshold",
			f:    reflect.ValueOf(apigw.GetLatencyAboveThresholdPerMin),
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
