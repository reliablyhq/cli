package scan

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/icza/dyno"
	finder "github.com/reliablyhq/cli/core/find"
	"github.com/reliablyhq/cli/core/terraform"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
)

func TestPolicyFind(t *testing.T) {
	t.Parallel()
	var p policy
	assert.NoError(t, p.find("terraform", "aws_autoscaling_group"))
	assert.Equal(t, ".reliably/policies/terraform/aws_autoscaling_group.rego", p.filepath)
	assert.Equal(t, "https://static.reliably.com/opa/terraform/aws_autoscaling_group.rego", p.uri)

	// check header
	headers := p.packageHeaders()
	assert.Len(t, headers, 3, "expected header length: 3")
	t.Logf("headers found --> %s", headers)
}

func TestTerraformEvaluate(t *testing.T) {
	t.Parallel()

	// read in terraform plan file
	b, err := os.ReadFile("../tests/terraform/plan.json")
	assert.NoError(t, err)

	var tfplan terraform.PlanRepresentation
	assert.NoError(t, json.Unmarshal(b, &tfplan))
	// testTarget.Item = tfplan

	// iterate module
	var targets []*Target
	for _, resource := range tfplan.PlannedValues.RootModule.Resources {
		targets = append(targets, &Target{
			ResourceType: resource.Type,
			Platform:     "terraform",
			Item:         resource,
		})
	}

	for _, resource := range tfplan.PlannedValues.RootModule.Resources {

		// get policy
		var p policy
		if resource.Type == "aws_launch_template" {
			assert.Error(t, p.find("terraform", resource.Type), "err expected: %s", ErrNotFound)
			continue
		}

		result, err := FindPolicyAndEvaluate(NewTarget(resource, "terraform", resource.Type))
		assert.NoError(t, err)
		assert.Len(t, result.Violations, 1, "expected violation length: 1")
		t.Logf("violations detected: %d", len(result.Violations))
		// for _, v := range result.Violations {
		// 	t.Logf("Expected Violation(s) Detected:\n%s\n----", v)
		// }
	}

}

func TestKubernetesEvaluate(t *testing.T) {
	t.Parallel()
	resources := finder.ReadAndSplitKubernetesFile("../tests/manifests/deployment.yaml")
	for _, r := range resources {
		header, err := finder.GetYamlInfo(r)
		if err != nil {
			// Unable to identify Yaml as K8s resource
			continue
		}

		var input interface{}
		assert.NoError(t, yaml.Unmarshal([]byte(r), &input))
		input = dyno.ConvertMapI2MapS(input)
		target := NewTarget(input, "kubernetes", header.Kind).AddSubGrouping(header.APIVersion)
		result, err := FindPolicyAndEvaluate(target)
		assert.NoError(t, err)

		assert.Len(t, result.Violations, 6, "expected violations: 6")
		t.Logf("violations detected: %d", len(result.Violations))
		// for _, v := range result.Violations {
		// 	t.Logf("Expected Violation(s) Detected:\n%s\n----", v)
		// }
	}

}
