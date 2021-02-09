package workflows

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWorkflowDir(t *testing.T) {

	var p string

	p = ".gitlab-ci.yml"
	d := workflowDir(p)
	assert.Equal(t, "", d, "Unexpected workflow folder")

	p = ".circleci/config.yml"
	d = workflowDir(p)
	assert.Equal(t, ".circleci", d, "Unexpected workflow folder")

	p = ".github/workflows/reliably.yaml"
	d = workflowDir(p)
	assert.Equal(t, ".github/workflows", d, "Unexpected workflow folder")
}
