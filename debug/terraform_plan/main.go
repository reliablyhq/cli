package main

import (
	"encoding/json"
	"os"

	"github.com/reliablyhq/cli/core"
	"github.com/reliablyhq/cli/core/terraform"
	log "github.com/sirupsen/logrus"
)

func main() {
	data, err := os.ReadFile("tfplan.json")
	if err != nil {
		log.Fatal(err)
	}

	var x terraform.PlanRepresentation
	if err := json.Unmarshal(data, &x); err != nil {
		log.Fatal(err)
	}

	policy, _ := os.ReadFile("policy.rego")

	resultSet := core.Eval(core.RegoModule{Name: "policy.rego", Raw: string(policy)}, &x)

	log.Printf("Results %v\n", len(resultSet))

	for _, result := range resultSet {
		log.Printf("Bindings: %v\n", len(result.Bindings))

		for _, binding := range result.Bindings {
			log.Printf("\t%v\n", binding)
		}

		log.Println("---")
		log.Printf("Expressions: %v\n", len(result.Expressions))

		var printFn func(map[string]interface{}, string)
		printFn = func(m map[string]interface{}, prefix string) {
			for k, v := range m {
				if x, isMap := v.(map[string]interface{}); isMap {
					log.Printf("%s %s\n", prefix, k)
					printFn(x, prefix+" ")
				} else {
					log.Printf("%s %s: %v\n", prefix, k, v)
				}
			}
		}

		for _, expr := range result.Expressions {
			m := expr.Value.(map[string]interface{})
			printFn(m, "")
		}
	}
}
