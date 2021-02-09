package workflows

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadGitlabTemplate(t *testing.T) {
	tpl, err := loadWorkflow(string(gitlab_Template))
	assert.Equal(t, nil, err, "Unexpected error")
	assert.NotEqual(t, nil, tpl, "Template is not properly loaded")
}

func TestEditGitlabWorkflow(t *testing.T) {
	var (
		wf  workflow
		tpl workflow
		err error
	)

	tpl, _ = loadWorkflow(gitlab_Template)

	// #################
	wf, _ = loadWorkflow(``)
	err = insertReliablyToGitlab(wf, tpl)
	assert.Equal(t, nil, err, "Unexpected error")
	assert.Equal(t, tpl, wf, "Editing an empty workflow shall be identical to template")

	// #################
	wf, _ = loadWorkflow(`
stages:
  - build
  - test

build-code-job:
  stage: build
  script:
    - echo "Build source code"

test-code-job1:
  stage: test
  script:
    - echo "Test job #1"

test-code-job2:
  stage: test
  script:
    - echo "Test job #2"
`)
	err = insertReliablyToGitlab(wf, tpl)
	assert.Equal(t, nil, err, "Unexpected error")
	_, hasCodeQuality := wf["code_quality"]
	assert.Equal(t, true, hasCodeQuality, "Missing code_quality")

	// #################
	wf, _ = loadWorkflow(`
stages:
  - qa

code_quality:
  stage: qa
  script:
    - echo "Another code quality job"
`)
	err = insertReliablyToGitlab(wf, tpl)
	assert.NotEqual(t, nil, err, "We don't edit the workflow if a 'code_quality' job already exists")

	// #################
	wf, _ = loadWorkflow(`
include: '.gitlab-ci-production.yml'
`)
	err = insertReliablyToGitlab(wf, tpl)
	assert.NotEqual(t, nil, err, "We don't edit the workflow if include is not a list")

	// #################
	wf, _ = loadWorkflow(`
include:
  - '.gitlab-ci-production.yml'
`)
	err = insertReliablyToGitlab(wf, tpl)
	assert.Equal(t, nil, err, "Unexpected error")
	includes := wf["include"].([]interface{})
	assert.Equal(t, 2, len(includes), "Include Code Quality template was not appended to list")

	// #################
	wf, _ = loadWorkflow(`
stages:
  - test

test-code-job:
  stage: test
  script:
    - echo "Simple test job"
`)
	err = insertReliablyToGitlab(wf, tpl)
	assert.Equal(t, nil, err, "Unexpected error")
	stages := wf["stages"].([]interface{})
	assert.Equal(t, 1, len(stages), "Test stage was already in list")

	// #################
	wf, _ = loadWorkflow(`
stages:
  - build

build-code-job:
  stage: build
  script:
    - echo "Build source code"
`)
	err = insertReliablyToGitlab(wf, tpl)
	assert.Equal(t, nil, err, "Unexpected error")
	stages = wf["stages"].([]interface{})
	assert.Equal(t, 2, len(stages), "Test stage was not appended to list")

}
