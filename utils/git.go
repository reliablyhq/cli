package utils

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	neturl "net/url"
	"os/exec"
	"runtime"
	"strings"

	"github.com/cli/safeexec"
)

var GitCommand = func(args ...string) (*exec.Cmd, error) {
	gitExe, err := safeexec.LookPath("git")
	if err != nil {
		programName := "git"
		if runtime.GOOS == "windows" {
			programName = "Git for Windows"
		}
		return nil, fmt.Errorf("unable to find git executable in PATH; please install %s before retrying", programName)
	}
	return exec.Command(gitExe, args...), nil
}

// ToplevelDir returns the top-level directory path of the current repository
func ToplevelDir() (string, error) {
	showCmd, err := GitCommand("rev-parse", "--show-toplevel")
	if err != nil {
		return "", err
	}
	output, err := showCmd.Output()
	return firstLine(output), err
}

// GitDir returns the directory path of the current repository
func GitDir() (string, error) {
	showCmd, err := GitCommand("rev-parse", "--git-dir")
	if err != nil {
		return "", err
	}
	output, err := showCmd.Output()
	return firstLine(output), err
}

// IsGitRepo indicates whether the current working directory is a git repository
func IsGitRepo() bool {
	d, err := GitDir()
	if err != nil {
		return false
	}

	return d != ""
}

func firstLine(output []byte) string {
	if i := bytes.IndexAny(output, "\n"); i >= 0 {
		return string(output)[0:i]
	}
	return string(output)
}

func GitRemoteOriginURL() (string, error) {
	showCmd, err := GitCommand("config", "--get", "remote.origin.url")
	if err != nil {
		return "", err
	}
	output, err := showCmd.Output()
	return firstLine(output), err
}

// ExtractOwnerRepoFromGitURL extracts owner and repo from a Git URL:
// https or ssh url
func ExtractOwnerRepoFromGitURL(url string) (owner string, repo string, err error) {

	if strings.HasSuffix(url, ".git") {
		url = strings.TrimSuffix(url, ".git")
	}

	if strings.HasPrefix(url, "https://") ||
		strings.HasPrefix(url, "http://") ||
		strings.HasPrefix(url, "git://") ||
		strings.HasPrefix(url, "ssh://") {
		p := strings.Split(url, "/")
		if len(p) >= 4 {
			owner, repo = p[3], p[4]
			return
		}
	}

	if strings.HasPrefix(url, "git@") {
		p := strings.Split(url, ":")
		p = strings.Split(p[1], "/")
		if len(p) >= 1 {
			owner, repo = p[0], p[1]
			return
		}
	}

	err = fmt.Errorf("Unable to extract owner/repo from %s", url)
	return
}

type GitRemoteURL struct {
	Proto string // https , ssh , git
	User  string // optional
	Host  string // host or host:port
	Path  string // full path to the repo (without .git ext)
	raw   string // raw url string
}

// Hash computes the hash for a git remote URL
// only on non changing fields: different protocol or user can still refer
// to the same remote origin
func (grURL *GitRemoteURL) Hash() string {
	str := fmt.Sprintf("%s:%s", grURL.Host, grURL.Path)
	h := sha256.New()
	_, _ = h.Write([]byte(str))
	return fmt.Sprintf("%x", h.Sum(nil))
}

// String returns the original raw URL of the git remote
func (grURL *GitRemoteURL) String() string {
	return grURL.raw
}

// ParseGitRemoteOriginURL parses a git remote origin url into
// a struct that can be uniquely identified regardless the used protocol
func ParseGitRemoteOriginURL(url string) (*GitRemoteURL, error) {
	raw := url

	if strings.HasSuffix(url, "/") {
		url = strings.TrimSuffix(url, "/")
	}
	if strings.HasSuffix(url, ".git") {
		url = strings.TrimSuffix(url, ".git")
	}

	if !strings.HasPrefix(url, "http://") &&
		!strings.HasPrefix(url, "https://") &&
		!strings.HasPrefix(url, "ssh://") &&
		!strings.HasPrefix(url, "git://") {
		// when no prefix is found, we consider scp-like syntax
		// for the SSH protocol: [user@]server:/path

		var (
			h string
			p string
		)

		split := strings.Split(url, ":")
		h = split[0]
		p = split[1]

		if strings.Contains(h, "@") {
			split = strings.Split(h, "@")
			h = split[1]
		}

		return &GitRemoteURL{
			raw:  raw,
			Host: h,
			Path: p,
		}, nil

	} else {

		p, err := neturl.Parse(url)
		if err != nil {
			return nil, err
		}

		return &GitRemoteURL{
			raw:   raw,
			Proto: p.Scheme,
			User:  p.User.Username(),
			Host:  p.Host,
			Path:  p.Path,
		}, nil

	}

}
