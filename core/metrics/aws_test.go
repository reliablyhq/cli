package metrics

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/stretchr/testify/assert"
)

func TestExtratArnFromResourceID(t *testing.T) {

	null := arn.ARN{}

	tests := []struct {
		name    string
		resID   string
		wantErr bool
	}{
		{
			name:    "missing ARN value",
			resID:   "aws/",
			wantErr: true,
		},
		{
			name:    "invalid ARN",
			resID:   "aws/this is not an arn",
			wantErr: true,
		},
		{
			name:    "incorrect ARN format",
			resID:   "aws/arn:aws:invalid",
			wantErr: true,
		},
		{
			name:    "valid ARN with aws/ prefix",
			resID:   "aws/arn:aws:rds:eu-west-1:123456789012:db:mysql-db",
			wantErr: false,
		},
		{
			name:    "ARN without aws/ prefix",
			resID:   "arn:aws:rds:eu-west-1:123456789012:db:mysql-db",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			arn, err := extractArnFromResourceID(tt.resID)

			if tt.wantErr {
				assert.NotEqual(t, nil, err, "Expected error not returned in result")
				assert.Equal(t, null, arn, "ARN should be zero value")
				return
			}

			assert.NoError(t, err, "Unexpected error")
			assert.NotEqual(t, null, arn, "ARN should not be zero value")
			assert.NotEqual(t, nil, arn, "ARN should not be nil")

		})
	}

}
