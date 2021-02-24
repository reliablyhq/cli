package core

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/open-policy-agent/opa/rego"
	log "github.com/sirupsen/logrus"

	"github.com/reliablyhq/cli/utils"
)

// Eval ...
// input requires to be a map !! with only string as keys !!
func Eval(policy string, input interface{}) rego.ResultSet {
	log.Debug(fmt.Sprintf("Evaluate policy %v", policy))

	ctx := context.Background()

	// Construct a Rego object that can be prepared or evaluated.
	r := rego.New(
		rego.Query("data"),
		rego.Load([]string{policy}, nil))

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
					)

					switch violation := violation.(type) {
					case string:
						ruleID = ""
						ruleDef = ""
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
					default:
						msg = "N/A"
					}

					res := Result{
						Resource: &resource,
						Rule:     Rule{ID: ruleID, Definition: ruleDef, Level: ruleLevel},
						Location: Location{e.Location.Row + startLine, e.Location.Col},
						Message:  msg}

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
