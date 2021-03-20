package core

import (
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/icza/dyno"
	"gopkg.in/yaml.v2"

	"github.com/reliablyhq/cli/utils"
)

func TestEval(t *testing.T) {

	manifest := filepath.Join("..", "tests", "manifests", "ns.yaml")
	policyPath := filepath.Join("..", "tests", "kubernetes.rego")
	policy, _ := ioutil.ReadFile(policyPath)

	// load manifest from file and unmarshall yaml
	var input interface{}

	fileContent, _ := ioutil.ReadFile(manifest)
	_ = yaml.Unmarshal([]byte(fileContent), &input)
	input = dyno.ConvertMapI2MapS(input)

	rmod := RegoModule{Name: "kubernetes.rego", Raw: string(policy)}

	rs := Eval(rmod, input)

	violations, _ := utils.NestedMapLookup(rs[0].Expressions[0].Value.(map[string]interface{}), "kubernetes", "violations")

	if violations != nil {
		vCount := len(violations.([]interface{}))
		t.Log("len of OPA eval violations", vCount)
		if vCount == 0 {
			t.Error("Policy eval did not find expected violations")
		}

	} else {
		t.Error("Policy eval did not find expected violations")
	}

}
