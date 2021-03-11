// Package scan contains...
package scan

import (
	"context"
	"errors"
	"fmt"

	"github.com/open-policy-agent/opa/rego"
	"github.com/prometheus/common/log"
)

var errNotImplemented = errors.New("not implemented")

// FindPolicyAndEvaluate -
func FindPolicyAndEvaluate(target *EvalTarget) (*EvalResult, error) {

	policy, err := findpolicy(target.Platform, target.ResourceType)
	if err != nil {
		return nil, err
	}

	return evaluate(policy, target.Item)
}

func findpolicy(platform, resourceType string) (*policy, error) {
	return nil, errNotImplemented
}

func findPolicyFuzzy(platform string, metadata map[string]string) ([]*policy, error) {
	return nil, errNotImplemented
}

func evaluate(p *policy, target *EvalTarget) (*EvalResult, error) {
	log.Debug(fmt.Sprintf("Evaluate policy %s", p))

	ctx := context.Background()

	// Construct a Rego object that can be prepared or evaluated.
	r := rego.New(
		rego.Query("data"),
		rego.Load([]string{p.filepath}, nil),
	)

	// Create a prepared query that can be evaluated.
	query, err := r.PrepareForEval(ctx)
	if err != nil {
		log.Debug("fatal #1")
		log.Fatal(err)
	}

	// Execute the prepared query.
	rs, err := query.Eval(ctx, rego.EvalInput(target.Item))
	if err != nil {
		log.Debug("fatal #2")
		log.Fatal(err)
	}

	extractResults(rs, target.Platform, target.ResourceType)

	return nil, errNotImplemented
}

// NestedMapLookup ...
// m:  a map from strings to other maps or values, of arbitrary depth
// ks: successive keys to reach an internal or leaf node (variadic)
// If an internal node is reached, will return the internal map
//
// Returns: (Exactly one of these will be nil)
// rval: the target node (if found)
// err:  an error created by fmt.Errorf
//
// https://gist.github.com/ChristopherThorpe/fd3720efe2ba83c929bf4105719ee967
// Licensed under the CC by 4.0 https://creativecommons.org/licenses/by/4.0/
//
func extractResults(result *rego.Result, ks ...string) (rval interface{}, err error) {
	var ok bool

	if len(ks) == 0 { // degenerate input
		return nil, fmt.Errorf("NestedMapLookup needs at least one key")
	}
	if rval, ok = m[ks[0]]; !ok {
		return nil, fmt.Errorf("key not found; remaining keys: %v", ks)
	} else if len(ks) == 1 { // we've reached the final key
		return rval, nil
	} else if m, ok = rval.(map[string]interface{}); !ok {
		return nil, fmt.Errorf("malformed structure at %#v", rval)
	} else { // 1+ more keys
		return NestedMapLookup(m, ks[1:]...)
	}
}
