package aws

import (
	"context"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
	"github.com/stretchr/testify/assert"
)

const (
	oneDay            = 24 * time.Hour
	arnInvalidService = "arn:partition:service:region:account-id:resource-type:resource-id"
	arnApiGateway     = "arn:aws:apigateway:eu-west-1::/apis/abcdef1234"
	arnELB            = "arn:aws:elasticloadbalancing:eu-west-1:123456789:loadbalancer/app/dummy/12az34er56"
)

var (
	now time.Time = time.Now()
)

func TestParseResourceID(t *testing.T) {

	noarn := arn.ARN{}
	zerovalue := AwsResource{arn: noarn}

	tests := []struct {
		name    string
		resID   string
		wantErr bool
	}{
		{
			name:    "missing resource ID",
			resID:   "",
			wantErr: true,
		},
		{
			name:    "invalid ARN",
			resID:   "this is not an arn",
			wantErr: true,
		},
		{
			name:    "incorrect ARN format",
			resID:   "arn:aws:invalid",
			wantErr: true,
		},
		{
			name:    "valid ARN with unsupported Service",
			resID:   "arn:aws:rds:eu-west-1:123456789012:db:mysql-db",
			wantErr: true,
		},
		{
			name:    "valid ARN",
			resID:   "arn:aws:apigateway:eu-west-2::/apis/trj7cyiqib",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			r, err := parseResourceID(tt.resID)

			if tt.wantErr {
				assert.NotEqual(t, nil, err, "Expected error not returned in result")
				assert.Equal(t, zerovalue, r, "AwsResource should be zero value")
				t.Log(err)
				return
			}

			assert.NoError(t, err, "Unexpected error")
			assert.NotEqual(t, zerovalue, r, "ARN should not be zero value")
		})
	}

}

func TestGetAwsResourceErrorRateInput(t *testing.T) {

	to := time.Now()
	from := to.Add(-oneDay)

	tests := []struct {
		name    string
		arn     string
		wantErr bool
	}{
		{
			name:    "Invalid AWS Service",
			arn:     arnInvalidService,
			wantErr: true,
		},
		{
			name:    "Api Gateway",
			arn:     arnApiGateway,
			wantErr: false,
		},
		{
			name:    "Application Load Balancer",
			arn:     arnELB,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			arn, _ := arn.Parse(tt.arn)
			r := AwsResource{arn: arn}

			provider, err := r.MetricProvider()
			if tt.wantErr {
				assert.NotEqual(t, nil, err, "Expected error not returned in result")
				t.Log(err)
				return
			}

			t.Log("provider", provider, "error", err)

			data, err := provider.GetErrorRateMetricDataInput(arn, from, to)
			assert.NoError(t, err, "Unexpected error")
			assert.NotEqual(t, nil, data, "metric data input is missing")
		})
	}

}

func TestGetAwsResourceLatencyInput(t *testing.T) {

	to := time.Now()
	from := to.Add(-oneDay)

	tests := []struct {
		name    string
		arn     string
		wantErr bool
	}{
		{
			name:    "Invalid AWS Service",
			arn:     arnInvalidService,
			wantErr: true,
		},
		{
			name:    "Api Gateway",
			arn:     arnApiGateway,
			wantErr: false,
		},
		{
			name:    "Application Load Balancer",
			arn:     arnELB,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			arn, _ := arn.Parse(tt.arn)
			r := AwsResource{arn: arn}

			provider, err := r.MetricProvider()
			if tt.wantErr {
				assert.NotEqual(t, nil, err, "Expected error not returned in result")
				t.Log(err)
				return
			}

			data, err := provider.GetLatencyMetricDataInput(arn, from, to)
			assert.NoError(t, err, "Unexpected error")
			assert.NotEqual(t, nil, data, "metric data input is missing")
		})
	}

}

func TestAwsProviderGetP99LatencyMetric(t *testing.T) {

	arn := arnApiGateway
	to := time.Now()
	from := to.Add(-oneDay)

	defer func() {
		delete(cloudwatchClients, "eu-west-1")
	}()

	tests := []struct {
		name       string
		mockedData *cloudwatch.GetMetricDataOutput
		want       float64
		wantErr    bool
	}{
		{
			name:    "no latency value retrieved",
			wantErr: true,
		},
		{
			name: "single latency value retrieved",
			mockedData: &cloudwatch.GetMetricDataOutput{
				MetricDataResults: []types.MetricDataResult{
					{
						Id:         aws.String("latency_percentile"),
						Timestamps: []time.Time{now},
						Values:     []float64{275.34},
					},
				},
			},
			want:    275.34,
			wantErr: false,
		},
		{
			name: "multiple latency values retrieved",
			mockedData: &cloudwatch.GetMetricDataOutput{
				MetricDataResults: []types.MetricDataResult{
					{
						Id:         aws.String("latency_percentile"),
						Timestamps: []time.Time{now, now, now},
						Values:     []float64{123, 456, 789},
					},
				},
			},
			want:    123,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			cloudwatchClients["eu-west-1"] = NewCloudWatchClientMock(tt.mockedData)

			cw := AwsCloudWatch{}
			p99, err := cw.Get99PercentLatencyMetricForResource(arn, from, to)

			if tt.wantErr {
				assert.NotEqual(t, nil, err, "Expected error not returned in result")
				t.Log(err)
				return
			}

			assert.NoError(t, err, "Unexpected error")
			assert.Equal(t, tt.want, p99, "Latency is not expected value")
		})
	}

}

func TestAwsProviderGetLatencyAboveThresholdMetric(t *testing.T) {

	arn := arnApiGateway
	to := time.Now()
	from := to.Add(-oneDay)
	threshold := 250

	defer func() {
		delete(cloudwatchClients, "eu-west-1")
	}()

	tests := []struct {
		name       string
		mockedData *cloudwatch.GetMetricDataOutput
		want       float64
		wantErr    bool
	}{
		{
			name:    "no latency value retrieved",
			wantErr: true,
		},
		{
			name: "single latency value retrieved - above threshold",
			mockedData: &cloudwatch.GetMetricDataOutput{
				MetricDataResults: []types.MetricDataResult{
					{
						Id:         aws.String("latency_above_threshold_per_min"),
						Timestamps: []time.Time{now},
						Values:     []float64{1},
					},
				},
			},
			want:    100,
			wantErr: false,
		},
		{
			name: "multiple latency values per minute (bool over threshold)",
			mockedData: &cloudwatch.GetMetricDataOutput{
				MetricDataResults: []types.MetricDataResult{
					{
						Id:         aws.String("latency_above_threshold_per_min"),
						Timestamps: []time.Time{now, now, now, now, now},
						Values:     []float64{1.0, 0.0, 1.0, 1.0, 0.0},
					},
				},
			},
			want:    60.0,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			cloudwatchClients["eu-west-1"] = NewCloudWatchClientMock(tt.mockedData)

			cw := AwsCloudWatch{}
			lp, err := cw.GetLatencyAboveThresholdPercentage(arn, from, to, threshold)

			if tt.wantErr {
				assert.NotEqual(t, nil, err, "Expected error not returned in result")
				t.Log(err)
				return
			}

			assert.NoError(t, err, "Unexpected error")
			assert.Equal(t, tt.want, lp, "Latency percentage is not expected value")
		})
	}

}

func TestAwsProviderGetErrorRateMetric(t *testing.T) {

	arn := arnApiGateway
	to := time.Now()
	from := to.Add(-oneDay)

	defer func() {
		delete(cloudwatchClients, "eu-west-1")
	}()

	tests := []struct {
		name       string
		mockedData *cloudwatch.GetMetricDataOutput
		want       float64
		wantErr    bool
	}{
		{
			name:    "no error rate value retrieved",
			wantErr: true,
		},
		{
			name: "single error rate value retrieved",
			mockedData: &cloudwatch.GetMetricDataOutput{
				MetricDataResults: []types.MetricDataResult{
					{
						Id:         aws.String("error_rate_percent"),
						Timestamps: []time.Time{now},
						Values:     []float64{75.99},
					},
				},
			},
			want:    75.99,
			wantErr: false,
		},
		{
			name: "multiple error rate values retrieved",
			mockedData: &cloudwatch.GetMetricDataOutput{
				MetricDataResults: []types.MetricDataResult{
					{
						Id:         aws.String("error_rate_percent"),
						Timestamps: []time.Time{now, now, now},
						Values:     []float64{55, 75, 99},
					},
				},
			},
			want:    55,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			cloudwatchClients["eu-west-1"] = NewCloudWatchClientMock(tt.mockedData)

			cw := AwsCloudWatch{}
			errorRate, err := cw.GetErrorPercentageMetricForResource(arn, from, to)

			if tt.wantErr {
				assert.NotEqual(t, nil, err, "Expected error not returned in result")
				t.Log(err)
				return
			}

			assert.NoError(t, err, "Unexpected error")
			assert.Equal(t, tt.want, errorRate, "Error rate is not expected value")
		})
	}

}

type CloudWatchClientMock struct {
	cloudwatch.Client
	getMetricDataOutput func() *cloudwatch.GetMetricDataOutput
}

func NewCloudWatchClientMock(data *cloudwatch.GetMetricDataOutput) *CloudWatchClientMock {
	if data == nil {
		data = &cloudwatch.GetMetricDataOutput{}
	}
	m := &CloudWatchClientMock{
		getMetricDataOutput: func() *cloudwatch.GetMetricDataOutput {
			return data
		},
	}
	return m
}

func (m *CloudWatchClientMock) GetMetricData(
	ctx context.Context, params *cloudwatch.GetMetricDataInput,
	optFns ...func(*cloudwatch.Options)) (*cloudwatch.GetMetricDataOutput, error) {
	return m.getMetricDataOutput(), nil
}
