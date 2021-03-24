package metrics

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
)

// we support ARN with resource kind in it
// arn:partition:service:region:account-id:resource-type/resource-id
// arn:partition:service:region:account-id:resource-type:resource-id
// see: https://docs.aws.amazon.com/general/latest/gr/aws-arns-and-namespaces.html

type AwsCloudWatch struct {
	Provider
}

func NewAwsCloudWatch() (*AwsCloudWatch, error) {

	c := cloudwatch.Client{}
	c = c
	b := arn.IsARN("arn:partition:service:region:account-id:resource-type:resource-id")
	b = b

	fmt.Println("Is valid ARN ? ", b)

	return nil, errors.New("NewAws not implemented")

}

func (cw *AwsCloudWatch) Get99PercentLatencyMetricForResource(resourceID string, from, to time.Time) (float64, error) {

	return -1, errors.New("Get99PercentLatencyMetricForResource not implemented")
}

func (cw *AwsCloudWatch) GetErrorPercentageMetricForResource(resourceID string, from, to time.Time) (float64, error) {
	return -1, errors.New("GetErrorPercentageMetricForResource not implemented")
}

// extractArnFromResourceID returns the ARN subpart of a service resource ID
// ie aws/arn:partition:service:region:account-id:resource-type:resource-id
func extractArnFromResourceID(id string) (arn.ARN, error) {

	var arnStr string
	if strings.HasPrefix(id, "aws/") {
		arnStr = strings.SplitN(id, "/", 2)[1] // ID is aws/arn:aws:...
	} else {
		arnStr = id // ID is directly arn:aws:...
	}

	if arnStr == "" {
		return arn.ARN{}, errors.New("Missing ARN in resource identifier: aws/arn:...")
	}

	if !arn.IsARN(arnStr) {
		return arn.ARN{}, fmt.Errorf("'%s' is not a valid ARN", arnStr)
	}

	return arn.Parse(arnStr)
}
