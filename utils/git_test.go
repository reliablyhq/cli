package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTopLevelDir(t *testing.T) {
	d, err := ToplevelDir()
	assert.Equal(t, err, nil, "Unexpected error")
	assert.NotEqual(t, d, "", "Top level dir should not be empty")
}

func TestGitDir(t *testing.T) {
	d, err := GitDir()
	assert.Equal(t, err, nil, "Unexpected error")
	assert.NotEqual(t, d, "", "Git dir should not be empty")
}

func TestIsGitRepo(t *testing.T) {
	b := IsGitRepo()
	assert.Equal(t, b, true, "The CLI is a git repo !")
}
