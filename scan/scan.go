// Package scan contains...
package scan

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"net/http"

	"github.com/open-policy-agent/opa/rego"

	log "github.com/sirupsen/logrus"
)

var errNotImplemented = errors.New("not implemented")

const (
	workspace = ".reliably"
	policyURL = "https://static.reliably.com/opa/%s"
)

// FindPolicyAndEvaluate -
func FindPolicyAndEvaluate(target *EvalTarget) (*EvalResult, error) {

	var p policy
	if err := p.find(target.Platform, target.ResourceType); err != nil {
		return nil, err
	}

	return p.evaluate(target)
}

func (p *policy) find(path ...string) (err error) {
	// check whether policy is already in cache folder
	// or download it from GitHub
	// and returns its content

	p.uri = fmt.Sprintf(policyURL, filepath.Join(path...)) + ".rego"
	path = append([]string{workspace, "policies"}, path...)
	p.filepath = strings.ToLower(filepath.Join(path...)) + ".rego"

	if _, err := os.Stat(p.filepath); os.IsNotExist(err) {
		// policy is not yet in local cache
		return p.download()
	}

	return
}

// downloads a given policy
// into the .reliably local policies cache
func (p *policy) download() (err error) {

	// --- create policy dir locally
	_ = os.MkdirAll(filepath.Dir(p.filepath), 0700) // ensure to create sub-folders if not exist yet

	// --- download policy file
	client := &http.Client{
		Timeout: time.Second * 2,
	}

	// Get the data
	resp, err := client.Get(p.uri)
	if err != nil {
		return fmt.Errorf("error fetching policy via API: %s", err)
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("No file found at URL: %v", p.uri)
	}

	defer resp.Body.Close()

	// Create the file
	out, err := os.Create(p.filepath)
	if err != nil {
		return fmt.Errorf("error creating policy file in local cache: %s", err)
	}
	defer out.Close()

	// Write the body to file
	if _, err = io.Copy(out, resp.Body); err != nil {
		return fmt.Errorf("error writing policy file to local cache: %s", err)
	}

	return
}

func (p *policy) evaluate(target *EvalTarget) (*EvalResult, error) {
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

	for _, r := range rs {
		v, err := extractResults(&r, target.Platform, "aws", target.ResourceType)
		if err != nil {
			return nil, err
		}

		_ = v
	}
	// extractResults(rs, target.Platform, target.ResourceType)

	// extractResults(rs, target.Platform, target.ResourceType)

	return nil, errNotImplemented
}

func extractResults(result *rego.Result, ks ...string) (rval interface{}, err error) {
	if len(ks) == 0 { // degenerate input
		return nil, fmt.Errorf("NestedMapLookup needs at least one key")
	}

	// extractor ...
	// m:  a map from strings to other maps or values, of arbitrary depth
	// ks: successive keys to reach an internal or leaf node (variadic)
	// If an internal node is reached, will return the internal map
	//
	// Returns: (Exactly one of these will be nil)
	// rval: the target node (if found)
	// err:  an error created by fm t.Errorf
	var extractor func(m map[string]interface{}, ks ...string) (interface{}, error)
	extractor = func(m map[string]interface{}, ks ...string) (interface{}, error) {
		var ok bool
		if rval, ok = m[ks[0]]; !ok {
			return nil, fmt.Errorf("key not found; remaining keys: %v", ks)
		} else if len(ks) == 1 { // we've reached the final key
			return rval, nil
		} else if m, ok = rval.(map[string]interface{}); !ok {
			return nil, fmt.Errorf("malformed structure at %#v", rval)
		} else { // 1+ more keys
			return extractor(m, ks[1:]...)
		}
	}

	var results []interface{}
	for _, expr := range result.Expressions {
		if m, ok := expr.Value.(map[string]interface{}); ok {
			v, err := extractor(m, ks...)
			if err != nil {
				return nil, err
			}

			results = append(results, v)
		}
	}
	rval = result
	return
}

// TODO:

func findPolicyFuzzy(platform string, metadata map[string]string) ([]*policy, error) {
	return nil, errNotImplemented
}
