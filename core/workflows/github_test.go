package workflows

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadGithubTemplate(t *testing.T) {
	tpl, err := loadWorkflow(string(github_Template))
	assert.Equal(t, nil, err, "Unexpected error")
	assert.NotEqual(t, nil, tpl, "Template is not properly loaded")
}
