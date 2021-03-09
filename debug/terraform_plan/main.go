package main

import (
	"encoding/json"
	"io/ioutil"
	"log"

	"github.com/reliablyhq/cli/core"
	"github.com/reliablyhq/cli/core/terraform"
)

func main() {
	data, err := ioutil.ReadFile("tfplan.json")
	if err != nil {
		log.Fatal(err)
	}

	var x terraform.PlanRepresentation
	if err := json.Unmarshal(data, &x); err != nil {
		log.Fatal(err)
	}

	rs := core.Eval("policy.rego", &x)
	log.Print(rs)
}
