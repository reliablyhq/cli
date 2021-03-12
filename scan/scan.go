// Package scan contains...
package scan

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
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
func FindPolicyAndEvaluate(targets ...*Target) (offendingTargets []*Target, err error) {
	if len(targets) == 0 {
		return nil, errors.New("FindPolicyAndEvaluate requires atleast one [target] to be specified")
	}

	// use initial target to find policy
	// TODO: may be worth adding logic to handle target specific policy lookups
	var p policy
	target := targets[0]
	if err = p.find(target.Platform, target.ResourceType); err != nil {
		return
	}

	return p.evaluate(targets...)
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

func (p policy) String() string {
	s, _ := ioutil.ReadFile(p.filepath)
	return string(s)

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
		return err
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("GET %s: %s ", p.uri, resp.Status)
	}

	defer resp.Body.Close()

	// Create the file
	out, err := os.Create(p.filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Write the body to file
	if _, err = io.Copy(out, resp.Body); err != nil {
		return err
	}

	return
}

// packageHeader - reads package name from policy string
// to create evaluations keys.
// For eg. package terraform.aws.resource --> []string{ tarraform, aws, resource }
func (p *policy) packageHeaders() []string {
	matches := regexp.MustCompile(`(?i)package.(.*)\n`).FindAllStringSubmatch(p.String(), -1)
	return strings.Split(matches[0][1], ".")
}

// evaluate - executes policy evaluation against a given target(s)
// a list of offendingTargets is returns, i.e targets with violations
func (p *policy) evaluate(targets ...*Target) (offendingTargets []*Target, err error) {
	log.Debug(fmt.Sprintf("policy: %s, evaluating %d target(s)", p, len(targets)))

	fmt.Println(p.String())
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

	var w sync.WaitGroup
	var targetChan = make(chan *Target, 1000)
	for _, target := range targets {
		w.Add(1)
		go func(target *Target) {
			defer w.Done()
			defer func() {
				// TODO: better error handle
				if err != nil {
					log.Debugf("error evaluating target: %v - %s", target, err)
				}
			}()

			// Execute the prepared query.
			rs, err := query.Eval(ctx, rego.EvalInput(target.Item))
			if err != nil {
				return
			}

			violations, err := violationLookup(&rs, p.packageHeaders()...)
			// violations, err := violationLookup(&rs, target.Platform, "aws", target.ResourceType)
			if err != nil {
				return
			}

			if len(violations) > 0 {
				// fmt.Println(target)
				target.Result.Violations = append(target.Result.Violations, violations...)
				targetChan <- target
			}

		}(target)
	}

	w.Wait()
	close(targetChan)

	// add offending targets
	for target := range targetChan {
		offendingTargets = append(offendingTargets, target)
	}
	return
}

func violationLookup(rs *rego.ResultSet, ks ...string) (violations []Rule, err error) {
	if len(ks) == 0 { // degenerate input
		return nil, fmt.Errorf("resultlookup needs at least one key")
	}

	// add final key *violatiions*
	ks = append(ks, "violations")

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
		var rval interface{}
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

	// var results []interface{}
	for _, r := range *rs {
		for _, expr := range r.Expressions {
			if m, ok := expr.Value.(map[string]interface{}); ok {
				v, err := extractor(m, ks...)
				if err != nil {
					return nil, err
				}

				if data, ok := v.([]interface{}); ok {
					for _, d := range data {
						vRule := d.(map[string]interface{})
						b, _ := json.Marshal(&vRule)
						var rule Rule
						if err = json.Unmarshal(b, &rule); err != nil {
							return nil, err
						}
						violations = append(violations, rule)
					}
				}
			}
		}
	}

	return
}

// TODO:

func findPolicyFuzzy(platform string, metadata map[string]string) ([]*policy, error) {
	return nil, errNotImplemented
}
