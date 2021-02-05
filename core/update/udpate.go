package update

import (
	"context"
	"errors"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/google/go-github/v33/github"
	"github.com/hashicorp/go-version"
	"gopkg.in/yaml.v2"
)

var gitDescribeSuffixRE = regexp.MustCompile(`-\d+-g[a-f0-9]{7}$`)

// ReleaseInfo stores information about a release
type ReleaseInfo struct {
	Version string `json:"tag_name"`
	URL     string `json:"html_url"`
}

type StateEntry struct {
	CheckedForUpdateAt time.Time   `yaml:"checked_for_update_at"`
	LatestRelease      ReleaseInfo `yaml:"latest_release"`
}

// CheckForUpdate checks whether a new release is available on GitHub
func CheckForUpdate(client *http.Client, stateFilePath, repo string, currentVersion string) (*ReleaseInfo, error) {

	stateEntry, _ := getStateEntry(stateFilePath)
	if stateEntry != nil && time.Since(stateEntry.CheckedForUpdateAt).Hours() < 24 {
		return nil, nil
	}

	releaseInfo, err := getLatestReleaseInfo(client, repo)
	if err != nil {
		return nil, err
	}

	err = setStateEntry(stateFilePath, time.Now(), *releaseInfo)
	if err != nil {
		return nil, err
	}

	if versionGreaterThan(releaseInfo.Version, currentVersion) {
		return releaseInfo, nil
	}

	return nil, nil
}

func getStateEntry(stateFilePath string) (*StateEntry, error) {
	content, err := ioutil.ReadFile(stateFilePath)
	if err != nil {
		return nil, err
	}

	var stateEntry StateEntry
	err = yaml.Unmarshal(content, &stateEntry)
	if err != nil {
		return nil, err
	}

	return &stateEntry, nil
}

func setStateEntry(stateFilePath string, t time.Time, r ReleaseInfo) error {
	data := StateEntry{CheckedForUpdateAt: t, LatestRelease: r}
	content, err := yaml.Marshal(data)
	if err != nil {
		return err
	}
	_ = ioutil.WriteFile(stateFilePath, content, 0600)

	return nil
}

func versionGreaterThan(v, w string) bool {
	w = gitDescribeSuffixRE.ReplaceAllStringFunc(w, func(m string) string {
		// removes the git suffix (eg -37-g66bdb2f) from a dev version
		return ""
	})

	vv, ve := version.NewVersion(v)
	vw, we := version.NewVersion(w)

	return ve == nil && we == nil && vv.GreaterThan(vw)
}

func getLatestReleaseInfo(client *http.Client, repo string) (*ReleaseInfo, error) {
	if repo == "" {
		return nil, errors.New("Missing github repository as 'owner/repo'")
	}

	s := strings.Split(repo, "/")
	owner, repo := s[0], s[1]

	ghClient := github.NewClient(client)
	ctx := context.Background()
	release, _, err := ghClient.Repositories.GetLatestRelease(ctx, owner, repo)
	if err != nil {
		return nil, err
	}

	latest := &ReleaseInfo{
		Version: string(*release.TagName),
		URL:     string(*release.HTMLURL),
	}
	return latest, nil
}
