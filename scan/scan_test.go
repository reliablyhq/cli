package scan

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	// target
	testTarget = &EvalTarget{
		ResourceType: "aws_autoscaling_group",
		Platform:     "terraform",
	}
)

func TestPolicyFind(t *testing.T) {
	var p policy
	assert.NoError(t, p.find(testTarget.Platform, testTarget.ResourceType))
	assert.Equal(t, ".reliably/policies/terraform/aws_autoscaling_group.rego", p.filepath)
	assert.Equal(t, "https://static.reliably.com/opa/terraform/aws_autoscaling_group.rego", p.uri)
}
