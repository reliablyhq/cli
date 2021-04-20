package aws

import (
	"testing"

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
