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
	assert.Condition(t, func() bool {
		if url == "https://github.com/reliablyhq/cli" || url == "git@github.com:reliablyhq/cli.git" {
			return true
		} else {
			return false
		}
	}, "Unexpected remote origin url")
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
		{
			name:  "git protocol url",
			url:   "git://github.com/koke/grit.git",
			wants: []string{"koke", "grit"},
		},
		{
			name:  "ssh protocol url",
			url:   "ssh://me@myserver.com/owner/repo.git",
			wants: []string{"owner", "repo"},
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

func TestParseGitRemoteOriginURL(t *testing.T) {
	urls := [...]string{
		"ssh://user@host.xz:1234/path/to/repo.git/",
		"ssh://user@host.xz/path/to/repo.git/",
		"ssh://host.xz:1234/path/to/repo.git/",
		"ssh://host.xz/path/to/repo.git/",
		"ssh://user@host.xz/path/to/repo.git/",
		"ssh://host.xz/path/to/repo.git/",
		"ssh://user@host.xz/~user/path/to/repo.git/",
		"ssh://host.xz/~user/path/to/repo.git/",
		"ssh://user@host.xz/~/path/to/repo.git",
		"ssh://host.xz/~/path/to/repo.git",
		"user@host.xz:/path/to/repo.git/",
		"host.xz:/path/to/repo.git/",
		"user@host.xz:~user/path/to/repo.git/",
		"host.xz:~user/path/to/repo.git/",
		"user@host.xz:path/to/repo.git",
		"host.xz:path/to/repo.git",
		"git://host.xz/path/to/repo.git/",
		"git://host.xz/~user/path/to/repo.git/",
		"http://host.xz/path/to/repo.git/",
		"https://host.xz/path/to/repo.git/",
	}
	for _, url := range urls {
		gURL, err := ParseGitRemoteOriginURL(url)
		t.Log(gURL.Host, gURL.Path, gURL.Hash(), err)
		assert.Equal(t, nil, err, "Unexpected error")
		assert.NotEqual(t, nil, gURL, "Git Remote URL shall not be nil")
	}
}

func TestParsedServerFromGitURL(t *testing.T) {
	tests := []struct {
		name  string
		url   string
		wants string
	}{
		{
			name:  "https url",
			url:   "https://github.com/reliablyhq/cli",
			wants: "github.com",
		},
		{
			name:  "git clone url",
			url:   "git@gitlab.com:reliably/reliably-discovery-demo.git",
			wants: "gitlab.com",
		},
		{
			name:  "git clone url - without username",
			url:   "gitlab.com:reliably/reliably-discovery-demo.git",
			wants: "gitlab.com",
		},
		{
			name:  "ssh url",
			url:   "ssh://dmartin35@bitbucket.org/dmartin35/misc.git",
			wants: "bitbucket.org",
		},
		{
			name:  "ssh url without username",
			url:   "ssh://bitbucket.org/project.git",
			wants: "bitbucket.org",
		},
		{
			name:  "git protocol url",
			url:   "git://github.com/koke/grit.git",
			wants: "github.com",
		},
		{
			name:  "git protocol url",
			url:   "git://github.com/koke/grit.git",
			wants: "github.com",
		},

		{
			name:  "ssh url wiht custom server & port",
			url:   "ssh://login@server.com:12345/repository.git",
			wants: "server.com:12345",
		},

		{
			name:  "ssh url wiht custom server & port",
			url:   "me@myserver.example.com:repos/myrepo.git",
			wants: "myserver.example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gURL, err := ParseGitRemoteOriginURL(tt.url)
			assert.Equal(t, nil, err, "Unexpected error")
			assert.Equal(t, tt.wants, gURL.Host, "server not same as expected")
		})
	}
}
