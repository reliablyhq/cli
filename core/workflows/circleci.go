package workflows

import (
	"errors"

	"github.com/MakeNowJust/heredoc/v2"
)

const circleci_Path string = ".circleci/config.yml"

var circleci_Template string = heredoc.Doc(`
version: 2.1

jobs:
  scan:
    docker:
      - image: ghcr.io/reliablyhq/cli/cli:latest
        environment:
          RELIABLY_TOKEN: $RELIABLY_TOKEN
    working_directory: /home
    steps:
      - checkout # check out the code in the project directory
      - run: reliably scan kubernetes .

workflows:
  reliably:
    jobs:
      - scan
`)

var circleci_AccessTokenHelp string = `
You must define %s as an environment variable in your project settings:
https://circleci.com/docs/2.0/env-vars/#setting-an-environment-variable-in-a-project
`

// insertReliablyToCircleci modifies the base workflow with the Reliably
// template parts
// If it seems reliably is already in the base workflow, or cannot be added,
// an error is returned
func insertReliablyToCircleci(wf workflow, tpl workflow) error {

	if _, ok := wf["version"]; !ok {
		wf["version"] = tpl["version"]
	}

	if _, ok := wf["jobs"]; !ok {
		wf["jobs"] = make(map[interface{}]interface{})
	}
	if _, found := wf["jobs"].(map[interface{}]interface{})["scan"]; found {
		return errors.New("Your workflow already contains a 'scan' job")
	}
	jobs := tpl["jobs"].(map[interface{}]interface{})
	wf["jobs"].(map[interface{}]interface{})["scan"] = jobs["scan"]

	if _, ok := wf["workflows"]; !ok {
		wf["workflows"] = make(map[interface{}]interface{})
	}
	if _, found := wf["workflows"].(map[interface{}]interface{})["reliably"]; found {
		return errors.New("Your workflow already contains a 'reliably' workflow")
	}
	workflows := tpl["workflows"].(map[interface{}]interface{})
	wf["workflows"].(map[interface{}]interface{})["reliably"] = workflows["reliably"]

	return nil

}
