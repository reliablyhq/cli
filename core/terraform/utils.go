package terraform

import (
	"bufio"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/hashicorp/hcl"
	"github.com/reliablyhq/cli/core"
)

// ExtractResourcesFromPlan analysis a plan and returns a collection of resources
func ExtractResourcesFromPlan(plan *PlanRepresentation) ([]*core.Resource, error) {
	if plan == nil {
		return nil, errors.New("plan is nil")
	}

	resources := make([]*core.Resource, len(plan.ResourceChanges))

	for i, resChange := range plan.ResourceChanges {
		resources[i] = &core.Resource{
			File: core.File{
				Filepath: "UNKNOWN",
			},
			StartingLine: 0,
			Platform:     Platform,
			Kind:         resChange.Type,
			Name:         resChange.Name,
			URI:          resChange.Address,
		}
	}

	return resources, nil
}

// UnmarshalTF - unmarshals .tf files by recursively removing
// unsupported HCL lines, for eg. terraform functions. The theory being that
// we'll only be interested in policing values we can directly support
func UnmarshalTF(tf []byte, v interface{}) error {
	var re = regexp.MustCompile(`Unknown token\:.(\d{1,5}):`)

	if err := hcl.Unmarshal(tf, v); err != nil {
		if !re.MatchString(err.Error()) {
			return err
		}

		ln := re.FindAllStringSubmatch(err.Error(), -1)[0][1]
		badlineNo, _ := strconv.Atoi(ln)

		var updated strings.Builder
		var lineNo = 1
		scanner := bufio.NewScanner(strings.NewReader(string(tf)))

		// remove lineNo
		for scanner.Scan() {
			if lineNo != badlineNo {
				fmt.Fprintln(&updated, scanner.Text())
			}
			lineNo++
		}
		return UnmarshalTF([]byte(updated.String()), v)
	}
	return nil
}
