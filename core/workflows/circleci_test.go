package workflows

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadCircleciTemplate(t *testing.T) {
	tpl, err := loadWorkflow(string(circleci_Template))
	assert.Equal(t, nil, err, "Unexpected error")
	assert.NotEqual(t, nil, tpl, "Template is not properly loaded")
}

func TestEditCircleci(t *testing.T) {
	var (
		wf  workflow
		tpl workflow
		err error
	)

	tpl, _ = loadWorkflow(circleci_Template)

	// #################
	wf, _ = loadWorkflow(``)
	err = insertReliablyToCircleci(wf, tpl)
	assert.Equal(t, nil, err, "Unexpected error")
	assert.Equal(t, tpl, wf, "Editing an empty workflow shall be identical to template")

	// #################
	wf, _ = loadWorkflow(`version: 2.1
jobs:
  build:
    steps:
      - checkout
      - run: echo "this is the build job"
  test:
    steps:
      - checkout
      - run: echo "this is the test job"
workflows:
  build_and_test:
    jobs:
      - build
      - test`)
	err = insertReliablyToCircleci(wf, tpl)
	assert.Equal(t, nil, err, "Unexpected error")
	assert.NotEqual(t, nil, wf["jobs"].(map[interface{}]interface{})["discover"], "Missing discover job")
	assert.NotEqual(t, nil, wf["workflows"].(map[interface{}]interface{})["reliably"], "Missing reliably workflow")

	// #################
	wf, _ = loadWorkflow(`version: 2.1
jobs:
  discover:
    steps:
      - checkout
      - run: echo "this is the reliably discover"
`)
	err = insertReliablyToCircleci(wf, tpl)
	assert.NotEqual(t, nil, err, "We don't edit the workflow if a 'discover' job already exists")

	// #################
	wf, _ = loadWorkflow(`version: 2.1
workflows:
  reliably:
    jobs:
     - unexpected
`)
	err = insertReliablyToCircleci(wf, tpl)
	assert.NotEqual(t, nil, err, "We don't edit the workflow if a 'reliably' workflow already exists")
}
