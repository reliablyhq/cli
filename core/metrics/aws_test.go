package metrics

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/stretchr/testify/assert"
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
