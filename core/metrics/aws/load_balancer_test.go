package aws

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
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
