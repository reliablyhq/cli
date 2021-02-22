package utils

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTopLevelDir(t *testing.T) {
	d, err := ToplevelDir()
	assert.Equal(t, nil, err, "Unexpected error")
	assert.NotEqual(t, "", d, "Top level dir should not be empty")
}

func TestGitDir(t *testing.T) {
	d, err := GitDir()
	assert.Equal(t, nil, err, "Unexpected error")
	assert.NotEqual(t, "", d, "Git dir should not be empty")
}

func TestIsGitRepo(t *testing.T) {
	b := IsGitRepo()
	assert.Equal(t, true, b, "The CLI is a git repo !")
}

func TestGitRemoteOriginURl(t *testing.T) {
	url, err := GitRemoteOriginURL()
	assert.Equal(t, nil, err, "Unexpected error")
	b := strings.Contains(url, "github.com")
	assert.Equal(t, true, b, "Remote origin URL is not on github.com")
	assert.Equal(t, "https://github.com/reliablyhq/cli", url, "Unexpected remote origin url")
}

func TestExtractOwnerRepoFromGitURL(t *testing.T) {
	tests := []struct {
		name  string
		url   string
		wants []string
	}{
		{
			name:  "github https url",
			url:   "https://github.com/reliablyhq/cli",
			wants: []string{"reliablyhq", "cli"},
		},
		{
			name:  "github ssh url",
			url:   "git@github.com:reliablyhq/cli",
			wants: []string{"reliablyhq", "cli"},
		},
		{
			name:  "github https url with .git ext",
			url:   "https://github.com/reliablyhq/cli.git",
			wants: []string{"reliablyhq", "cli"},
		},
		{
			name:  "github ssh url with .git ext",
			url:   "git@github.com:reliablyhq/cli.git",
			wants: []string{"reliablyhq", "cli"},
		},
		{
			name:  "gitlab https url",
			url:   "https://gitlab.com/reliably/reliably-discovery-demo.git",
			wants: []string{"reliably", "reliably-discovery-demo"},
		},
		{
			name:  "gitlab ssh url",
			url:   "git@gitlab.com:reliably/reliably-discovery-demo.git",
			wants: []string{"reliably", "reliably-discovery-demo"},
		},

		{
			name:  "bitbucket https url",
			url:   "https://dmartin35@bitbucket.org/dmartin35/misc.git",
			wants: []string{"dmartin35", "misc"},
		},
		{
			name:  "bitbucket ssh url",
			url:   "git@bitbucket.org:dmartin35/misc.git",
			wants: []string{"dmartin35", "misc"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			owner, repo, err := ExtractOwnerRepoFromGitURL(tt.url)
			assert.Equal(t, nil, err, "Unexpected error")

			found := []string{owner, repo}
			assert.Equal(t, tt.wants, found, "owner/repo not same as expected")
		})
	}

}
