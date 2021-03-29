package aws

import (
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/stretchr/testify/assert"
)

const (
	oneDay            = 24 * time.Hour
	arnInvalidService = "arn:partition:service:region:account-id:resource-type:resource-id"
	arnApiGateway     = "arn:aws:apigateway:eu-west-1::/apis/abcdef1234"
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			arn, _ := arn.Parse(tt.arn)
			r := AwsResource{arn: arn}

			data, err := r.GetErrorRateMetricDataInput(from, to)

			if tt.wantErr {
				assert.NotEqual(t, nil, err, "Expected error not returned in result")
				t.Log(err)
				return
			}

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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			arn, _ := arn.Parse(tt.arn)
			r := AwsResource{arn: arn}

			t.Log(arn)
			t.Log(r)

			data, err := r.GetLatencyMetricDataInput(from, to)

			if tt.wantErr {
				assert.NotEqual(t, nil, err, "Expected error not returned in result")
				t.Log(err)
				return
			}

			assert.NoError(t, err, "Unexpected error")
			assert.NotEqual(t, nil, data, "metric data input is missing")
		})
	}

}
