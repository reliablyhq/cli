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

var (
	errNotImplemented = errors.New("not implemented")
	errNotFound       = errors.New("not found")
)

const (
	workspace = ".reliably"
	policyURL = "https://static.reliably.com/opa/%s"
)

// EvaluateMany policies concurrently - // TODO: WEP
func EvaluateMany(targets []*Target, resultChan chan *Result, errorChan chan error) {
	doWork := func(target *Target, rChan chan *Result, errChan chan error, wg *sync.WaitGroup) {
		defer wg.Done()
		r, e := FindPolicyAndEvaluate(target)

		if e != nil && errChan != nil {
			errorChan <- e
			return
		}

		if r != nil && rChan != nil {
			rChan <- r
		}
	}

	var wg sync.WaitGroup
	wg.Add(len(targets))
	for _, t := range targets {
		go doWork(t, resultChan, errorChan, &wg)
	}
	wg.Wait()
}

// FindPolicyAndEvaluate -
func FindPolicyAndEvaluate(target *Target) (result *Result, err error) {
	// if len(targets) == 0 {
	// 	return nil, errors.New("FindPolicyAndEvaluate requires atleast one [target] to be specified")
	// }

	// use initial target to find policy
	// TODO: may be worth adding logic to handle target specific policy lookups
	var p policy

	// build path
	path := []string{target.Platform}
	path = append(path, target.subgroups...)
	path = append(path, target.ResourceType)

	if err = p.find(path...); err != nil {
		return
	}

	return p.evaluate(target)
}

// NewTarget - returns a pointer an instance of scan.Target
func NewTarget(item interface{}, platform, resourcetype string) *Target {
	return &Target{
		Item:         item,
		ResourceType: resourcetype,
		Platform:     platform,
	}
}

// AddSubGrouping - adds a sub group
// subgroups are used by some resources that require additional partition beyond
// platform --> resource
// Note that subgroups are placed between platform & resource
// during lookups, i.e platform --> [subgroups] --> resource
// or kubernetes --> app/v1 --> resource
func (t *Target) AddSubGrouping(groups ...string) *Target {
	for _, g := range groups {
		t.subgroups = append(t.subgroups, g)
	}
	return t
}

// find - sets the path to the cached policy or
// downloads it if unavailable
func (p *policy) find(path ...string) (err error) {
	p.uri = strings.ToLower(fmt.Sprintf(policyURL, filepath.Join(path...)) + ".rego")
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
		if resp.StatusCode == 404 {
			return errNotFound
		}
		return fmt.Errorf("GET %s: %s", p.uri, resp.Status)
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
func (p *policy) evaluate(target *Target) (*Result, error) {
	// log.Debug(fmt.Sprintf("policy: %s, evaluating %d target(s)", p, len(targets)))
	var result Result

	// fmt.Println(p.String())
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
		return nil, err
	}

	violations, err := violationLookup(&rs, p.packageHeaders()...)
	// violations, err := violationLookup(&rs, target.Platform, "aws", target.ResourceType)
	if err != nil {
		return nil, err
	}

	result.Violations = violations

	return &result, nil
}

// violationLookup - handles looking up violations in targets against
// policies recursively.
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

	// Iterate expressions and evaluate violations (if any)
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
func (p *policy) findPolicyFuzzy(platform string, metadata map[string]string) ([]*policy, error) {
	return nil, errNotImplemented
}
