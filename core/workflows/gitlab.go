package workflows

import (
	"errors"
	"fmt"

	"github.com/MakeNowJust/heredoc/v2"

	"github.com/reliablyhq/cli/utils"
)

const gitlab_Path string = ".gitlab-ci.yml"

var gitlab_Template string = heredoc.Doc(`
stages:
  - test

include:
  - template: Code-Quality.gitlab-ci.yml

code_quality:
  image:
    name: ghcr.io/reliablyhq/cli/cli:latest
    entrypoint: ["/bin/sh", "-c"]
  script:
    - reliably scan kubernetes . --format codeclimate --output gl-code-quality-report.json || true
  stage: test
  allow_failure: true
  artifacts:
    when: always
    expose_as: 'Code Quality Report'
    paths: [gl-code-quality-report.json]
  rules:
    - if: $RELIABLY_TOKEN
`)

var gitlab_AccessTokenHelp string = `
You must define %s as a masked Variable in your project CI/CD settings:
https://gitlab.com/%s/%s/-/settings/ci_cd
`

// insertReliablyToGitlab modifies a Gitlab CI yaml file with the
// Reliably code quality parts
// If it cannot update some parts of the initial file, it returns an error
func insertReliablyToGitlab(wf workflow, tpl workflow) error {

	if _, ok := wf["code_quality"]; !ok {
		wf["code_quality"] = tpl["code_quality"]
	} else {
		return errors.New("Your workflow already contains a 'code_quality' job")
	}

	if _, ok := wf["stages"]; !ok {
		wf["stages"] = make([]interface{}, 0)
	}
	if _, ok := wf["stages"].([]interface{}); !ok {
		return errors.New("Unable to edit your worklow: 'stages' is not a valid list")
	}
	stage := fmt.Sprint(tpl["stages"].([]interface{})[0])
	wfStages := wf["stages"].([]interface{})
	stages := make([]string, len(wfStages))
	for i := range wfStages {
		stages[i] = fmt.Sprint(wfStages[i])
	}

	if !utils.StringInArray(stage, stages) {
		wf["stages"] = append(wfStages, stage)
	}

	if _, ok := wf["include"]; !ok {
		wf["include"] = make([]interface{}, 0)
	}
	if _, ok := wf["include"].([]interface{}); !ok {
		return errors.New("Unable to edit your worklow: 'include' is not a valid list")
	}
	wfIncludes := wf["include"].([]interface{})
	wf["include"] = append(wfIncludes, tpl["include"].([]interface{})[0])

	return nil
}
