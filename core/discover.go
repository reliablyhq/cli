package core

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/open-policy-agent/opa/rego"
	//"github.com/open-policy-agent/opa/storage/inmem"
	//opaUtil "github.com/open-policy-agent/opa/util"
	log "github.com/sirupsen/logrus"

	"github.com/reliablyhq/cli/utils"
)

var podPolicy = `package kubernetes.pod

violations[rule] {
  input.kind = "Pod"
  container = input.spec.containers[_]
  endswith(container.image, ":latest")

  rule := {
    "ruleID": "K8S-POD-0001",
    "ruleDef": "Image with latest tag",
    "level": 2,
    "message": "You should not use the default 'latest' image tag. It causes ambiguity and leads to the cluster not pulling the new image"
  }
}

violations[rule] {
  input.kind = "Pod"
  container = input.spec.containers[_]
  not contains(container.image, ":")

  rule := {
    "ruleID": "K8S-POD-0002",
    "ruleDef": "Missing image tag",
    "level": 1,
    "message": "It's best practice to always use image tags "
  }
}

violations[rule] {
  input.kind == "Pod"
  container = input.spec.containers[_]
  not contains(container.image, "/")

  rule := {
    "ruleID": "K8S-POD-0003",
    "ruleDef": "Image from non-approved source",
    "level": 2,
    "message": "Only images from an approved registry can be run"
  }
}`

var deployPolicy = `
package kubernetes.deployment

missing_replica {
	not input.spec.replicas
}

single_replica {
	input.spec.replicas == 1
}

has_requests_resource {
	input.spec.requests
}

has_limits_resource {
	input.spec.limits
}

has_high_requests_memory_resource {
    mem := input.spec.template.spec.containers[_].resources.requests.memory
    is_string(mem)
	mem == "512Mi"
}

has_high_requests_memory_resource {
    mem := input.spec.template.spec.containers[_].resources.requests.memory
    is_string(mem)
	mem == "1024Mi"
}

has_high_requests_memory_resource {
    mem := input.spec.template.spec.containers[_].resources.requests.memory
    is_string(mem)
	mem == "1Gi"
}

has_high_limits_memory_resource {
    mem := input.spec.template.spec.containers[_].resources.limits.memory
    is_string(mem)
	mem == "512Mi"
}

has_high_limits_memory_resource {
    mem := input.spec.template.spec.containers[_].resources.limits.memory
    is_string(mem)
	mem == "1024Mi"
}

has_high_limits_memory_resource {
    mem := input.spec.template.spec.containers[_].resources.limits.memory
    is_string(mem)
	mem == "1Gi"
}

has_high_requests_cpu_resource {
    cpu := input.spec.template.spec.containers[_].resources.requests.cpu
    is_number(cpu)
	cpu >= 0.8
}

has_high_requests_cpu_resource {
    cpu := input.spec.template.spec.containers[_].resources.requests.cpu
    is_string(cpu)
	cpu == "1"
}

has_high_requests_cpu_resource {
    cpu := input.spec.template.spec.containers[_].resources.requests.cpu
    is_string(cpu)
	cpu == "1000m"
}

has_high_limits_cpu_resource {
    cpu := input.spec.template.spec.containers[_].resources.limits.cpu
    is_number(cpu)
	cpu >= 0.8
}

has_high_limits_cpu_resource {
    cpu := input.spec.template.spec.containers[_].resources.limits.cpu
    is_string(cpu)
	cpu == "1"
}

has_high_limits_cpu_resource {
    cpu := input.spec.template.spec.containers[_].resources.limits.cpu
    is_string(cpu)
	cpu == "1000m"
}

missing_cpu_requests_resource {
    missing_cpu := {
        container | container := input.spec.template.spec.containers[_]
        not container.resources.requests.cpu
        not container.resources.limits.cpu
    }
    count(missing_cpu) > 0
}

missing_cpu_requests_but_has_cpu_limits_resource {
    missing_cpu := {
        container | container := input.spec.template.spec.containers[_]
        not container.resources.requests.cpu
        container.resources.limits.cpu
    }
    count(missing_cpu) > 0
}

image_pull_policy_set_to_always {
    image_policy := input.spec.template.spec.containers[_].imagePullPolicy
    image_policy == "Always"
}

missing_image_pull_policy  {
    missing_image_policy := {
        container | container := input.spec.template.spec.containers[_]
        not container.imagePullPolicy
    }
    count(missing_image_policy) > 0
}

missing_rollout_strategy {
    not input.spec.template.spec.strategy.rollingUpdate
}

missing_minready_seconds {
    not input.spec.template.spec.minReadySeconds
}

violations[rule] {
  input.kind == "Deployment"
  missing_replica
  msg := "You should specify a number of replicas"

  rule := {
    "ruleID": "K8S-DPL-0001",
    "ruleDef": "Missing replicas",
    "level": 2,
    "message": msg
  }
}

violations[rule] {
  input.kind == "Deployment"
  single_replica
  msg := "A single replica may lead to downtime"

  rule := {
    "ruleID": "K8S-DPL-0002",
    "ruleDef": "Single replica",
    "level": 1,
    "message": msg
  }
}

violations[rule] {
  input.kind == "Deployment"
  has_requests_resource
  msg := "Specifying resource requests means the Kubernetes scheduler will be able to fit your pod more appropriately and optimise resource usage"

  rule := {
    "ruleID": "K8S-DPL-0003",
    "ruleDef": "Missing resource requests",
    "level": 3,
    "message": msg
  }
}

violations[rule] {
  input.kind == "Deployment"
  has_requests_resource
  msg := "Specifying resource limits means the Pod will not be able to consume all the resources of the underlying node"

  rule := {
    "ruleID": "K8S-DPL-0004",
    "ruleDef": "Missing resource limits",
    "level": 3,
    "message": msg
  }
}

violations[rule] {
  input.kind == "Deployment"
  has_high_requests_memory_resource
  msg := "Setting a high memory request may render pod scheduling difficult"

  rule := {
    "ruleID": "K8S-DPL-0005",
    "ruleDef": "High memory requests",
    "level": 2,
    "message": msg
  }
}

violations[rule] {
  input.kind == "Deployment"
  has_high_limits_memory_resource
  msg := "Setting a high memory request may render pod scheduling difficult and/or starve other pods from memory on a node"

  rule := {
    "ruleID": "K8S-DPL-0006",
    "ruleDef": "High memory limits",
    "level": 2,
    "message": msg
  }
}

violations[rule] {
  input.kind == "Deployment"
  missing_cpu_requests_resource
  msg := "Setting a high cpu request may render pod scheduling difficult or starve other pods"

  rule := {
    "ruleID": "K8S-DPL-0007",
    "ruleDef": "Missing CPU requests",
    "level": 3,
    "message": msg
  }
}

violations[rule] {
  input.kind == "Deployment"
  has_high_requests_cpu_resource
  msg := "Setting a high cpu request may render pod scheduling difficult"

  rule := {
    "ruleID": "K8S-DPL-0008",
    "ruleDef": "High CPU requests",
    "level": 2,
    "message": msg
  }
}

violations[rule] {
  input.kind == "Deployment"
  missing_cpu_requests_resource
  msg := "Not setting a cpu requests means the pod will be allowed to consume the entire available CPU (unless the cluster has set a global limit)"

  rule := {
    "ruleID": "K8S-DPL-0009",
    "ruleDef": "Missing CPU requests",
    "level": 3,
    "message": msg
  }
}

violations[rule] {
  input.kind == "Deployment"
  missing_cpu_requests_but_has_cpu_limits_resource
  msg := "Setting uniquely a cpu limits and not a cpu requests means the pod will be allocated upto cpu limits by the scheduler"

  rule := {
    "ruleID": "K8S-DPL-0010",
    "ruleDef": "Improperly configured CPU requests & limits",
    "level": 3,
    "message": msg
  }
}

violations[rule] {
  input.kind == "Deployment"
  missing_image_pull_policy
  msg := "Not setting the image pull policy means it be set to 'Always' which is not recommended"

  rule := {
    "ruleID": "K8S-DPL-0011",
    "ruleDef": "Missing image pull policy",
    "level": 1,
    "message": msg
  }
}

violations[rule] {
  input.kind == "Deployment"
  image_pull_policy_set_to_always
  msg := "Image pull policy should usually not be set to 'Always'"

  rule := {
    "ruleID": "K8S-DPL-0012",
    "ruleDef": "Image pull policy shall not be always",
    "level": 1,
    "message": msg
  }
}

violations[rule] {
  input.kind == "Deployment"
  missing_rollout_strategy
  msg := "A rollout strategy can reduce the risk of downtime"

  rule := {
    "ruleID": "K8S-DPL-0013",
    "ruleDef": "Missing rollout strategy",
    "level": 2,
    "message": msg
  }
}

violations[rule] {
  input.kind == "Deployment"
  missing_minready_seconds
  msg := "Without the 'minReadySeconds' property set, pods are considered available from the first time the readiness probe is valid. Settings this value indicates how long it the pod should be ready for before being considered available."

  rule := {
    "ruleID": "K8S-DPL-0014",
    "ruleDef": "Missing minReady seconds",
    "level": 2,
    "message": msg
  }
}
`

type RegoModule struct {
	Name string
	Raw  string
}

// Eval ...
// input requires to be a map !! with only string as keys !!
func Eval(policy RegoModule, input interface{}) rego.ResultSet {
	log.Debug(fmt.Sprintf("Evaluate policy %s", policy.Name))

	ctx := context.Background()

	/*
		var json map[string]interface{}

		err := opaUtil.UnmarshalJSON([]byte(podPolicy), &json)
		if err != nil {
			// Handle error.
		}
	*/

	// Manually create the storage layer. inmem.NewFromObject returns an
	// in-memory store containing the supplied data.
	//store := inmem.NewFromObject(json)

	// Construct a Rego object that can be prepared or evaluated.
	r := rego.New(
		rego.Query("data"),
		//rego.Load([]string{policy}, nil))
		//rego.Store(store))
		//rego.Module("policy.rego", podPolicy),
		rego.Module(policy.Name, policy.Raw))

	// Create a prepared query that can be evaluated.
	query, err := r.PrepareForEval(ctx)
	if err != nil {
		log.Debug("fatal #1")
		log.Fatal(err)
	}

	// Execute the prepared query.
	rs, err := query.Eval(ctx, rego.EvalInput(input))
	if err != nil {
		log.Debug("fatal #2")
		log.Fatal(err)
	}

	return rs
}

// PrintViolations ...
// staringLine is needed for files that contains multiple split resources
// in order to be able to match the location to exact line in original file
func PrintViolations(rs rego.ResultSet, filename string, platform string, kind string, startLine int) {
	log.Debug("PrintViolations", rs, filename, platform, kind, startLine)
	lplatform := strings.ToLower(platform)
	lkind := strings.ToLower(kind)

	for _, r := range rs {
		for _, e := range r.Expressions {
			violations, _ := utils.NestedMapLookup(e.Value.(map[string]interface{}), lplatform, lkind, "violations")
			if violations != nil {
				for _, msg := range violations.([]interface{}) {
					fmt.Printf("%v:%v:%v: %v\n", filename, e.Location.Row+startLine, e.Location.Col, msg)
				}
			}
		}
	}

}

// PrintViolationsOnWriter writes violations on the given writer
func PrintViolationsOnWriter(writer *SafeWriter, rs rego.ResultSet, filename string, platform string, kind string, startLine int) {
	log.Debug("PrintViolations", rs, filename, platform, kind, startLine)
	lplatform := strings.ToLower(platform)
	lkind := strings.ToLower(kind)

	for _, r := range rs {
		for _, e := range r.Expressions {
			violations, _ := utils.NestedMapLookup(e.Value.(map[string]interface{}), lplatform, lkind, "violations")
			if violations != nil {
				for _, msg := range violations.([]interface{}) {
					l := fmt.Sprintf("%v:%v:%v: %v\n", filename, e.Location.Row+startLine, e.Location.Col, msg)
					writer.Writeln(l)
				}
			}
		}
	}

}

// CountViolations returns the count of violations from an OPA result set
func CountViolations(rs rego.ResultSet, platform string, kind string) int {

	lplatform := strings.ToLower(platform)
	lkind := strings.ToLower(kind)

	for _, r := range rs {
		for _, e := range r.Expressions {
			violations, _ := utils.NestedMapLookup(e.Value.(map[string]interface{}), lplatform, lkind, "violations")
			if violations != nil {
				return len(violations.([]interface{}))
			}
		}
	}

	return 0
}

// ReportViolations iterates over OPA rego ResultSet to return an slice of internal Result structure
func ReportViolations(rs rego.ResultSet, filename string, platform string, kind string, startLine int, name string, uri string) ResultSet {
	log.Debug("ReportViolations", rs, filename, platform, kind, startLine)
	lplatform := strings.ToLower(platform)
	lkind := strings.ToLower(kind)

	f := File{filename}
	resource := Resource{
		File:         f,
		StartingLine: startLine,
		Platform:     platform,
		Kind:         kind,
		Name:         name,
		URI:          uri,
	}

	var results ResultSet

	for _, r := range rs {
		for _, e := range r.Expressions {
			violations, _ := utils.NestedMapLookup(e.Value.(map[string]interface{}), lplatform, lkind, "violations")
			if violations != nil {
				for _, violation := range violations.([]interface{}) {

					var (
						msg       string
						ruleID    string = ""
						ruleDef   string = ""
						ruleLevel Level
						example   string = ""
					)

					switch violation := violation.(type) {
					case string:
						msg = violation
					case map[string]interface{}:
						v := violation
						if v["ruleID"] != nil {
							ruleID = v["ruleID"].(string)
						}
						if v["ruleDef"] != nil {
							ruleDef = v["ruleDef"].(string)
						}
						if v["level"] != nil {
							l, err := v["level"].(json.Number).Int64()
							if err == nil {
								ruleLevel = Level(uint(l))
							}
						}
						msg = v["message"].(string)
						if v["example"] != nil {
							example = v["example"].(string)
						}
					default:
						msg = "N/A"
					}

					res := Result{
						Resource: &resource,
						Rule:     Rule{ID: ruleID, Definition: ruleDef, Level: ruleLevel},
						Location: Location{e.Location.Row + startLine, e.Location.Col},
						Message:  msg,
						Example:  example,
					}

					results = append(results, res)
					log.Debug(
						fmt.Sprintf("> %v:%v:%v: %v\n", filename, e.Location.Row+startLine, e.Location.Col, msg))
				}
			}
		}
	}

	return results

}

// ConvertViolationsToSuggestions iterates over internal complex ResultSet
// to return a slice of Suggestion better suited for output reporting
func ConvertViolationsToSuggestions(rs ResultSet, live bool) []*Suggestion {

	var suggestions []*Suggestion

	for _, r := range rs {
		suggestions = append(suggestions, NewSuggestion(r, live))
	}

	return suggestions
}
