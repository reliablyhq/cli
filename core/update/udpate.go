package update

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
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

	if VersionGreaterThan(releaseInfo.Version, currentVersion) {
		return releaseInfo, nil
	}

	return nil, nil
}

func getStateEntry(stateFilePath string) (*StateEntry, error) {
	content, err := os.ReadFile(stateFilePath)
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
	_ = os.WriteFile(stateFilePath, content, 0600)

	return nil
}

func VersionGreaterThan(v, w string) bool {
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

func GetLatestRelease(client *http.Client, repo string) (*github.RepositoryRelease, error) {
	if repo == "" {
		return nil, errors.New("Missing github repository as 'owner/repo'")
	}

	s := strings.Split(repo, "/")
	owner, repo := s[0], s[1]

	ghClient := github.NewClient(client)
	ctx := context.Background()
	latest, _, err := ghClient.Repositories.GetLatestRelease(ctx, owner, repo)
	return latest, err
}

func GetLatestReleaseAsset(client *http.Client, repo string, goos string) (*github.ReleaseAsset, error) {

	r, err := GetLatestRelease(client, repo)
	if err != nil {
		return nil, err
	}

	var bin string = fmt.Sprintf("reliably-%s-amd64", goos)

	for _, asset := range r.Assets {
		if *asset.Name == bin {
			return asset, nil
		}
	}

	return nil, fmt.Errorf("No asset found for your platform '%s' in the latest release", goos)

}

func DownloadLatestReleaseAsset(client *http.Client, repo string, goos string) (io.ReadCloser, error) {
	a, err := GetLatestReleaseAsset(client, repo, goos)
	if err != nil {
		return nil, err
	}

	s := strings.Split(repo, "/")
	owner, repo := s[0], s[1]

	ghClient := github.NewClient(client)
	ctx := context.Background()

	if client == nil {
		// we want to follow redirect for downloading, requires an explicit http client
		client = http.DefaultClient
	}
	// return (rc io.ReadCloser, redirectURL string, err error)
	rc, _, err := ghClient.Repositories.DownloadReleaseAsset(ctx, owner, repo, *a.ID, client)

	return rc, err
}

func GetRelease(client *http.Client, repo string, tag string) (*github.RepositoryRelease, error) {
	if repo == "" {
		return nil, errors.New("Missing github repository as 'owner/repo'")
	}

	s := strings.Split(repo, "/")
	owner, repo := s[0], s[1]

	ghClient := github.NewClient(client)
	ctx := context.Background()
	r, _, err := ghClient.Repositories.GetReleaseByTag(ctx, owner, repo, tag)

	return r, err
}

func GetReleaseAsset(client *http.Client, repo string, goos string, tag string) (*github.ReleaseAsset, error) {

	r, err := GetRelease(client, repo, tag)
	if err != nil {
		return nil, err
	}

	var bin string = fmt.Sprintf("reliably-%s-amd64", goos)

	for _, asset := range r.Assets {
		if *asset.Name == bin {
			return asset, nil
		}
	}

	return nil, fmt.Errorf("No asset found for your platform '%s' in the release '%s'", goos, tag)

}

func DownloadReleaseAsset(client *http.Client, repo string, goos string, tag string) (io.ReadCloser, error) {
	a, err := GetReleaseAsset(client, repo, goos, tag)
	if err != nil {
		return nil, err
	}

	s := strings.Split(repo, "/")
	owner, repo := s[0], s[1]

	ghClient := github.NewClient(client)
	ctx := context.Background()

	if client == nil {
		// we want to follow redirect for downloading, requires an explicit http client
		client = http.DefaultClient
	}
	// return (rc io.ReadCloser, redirectURL string, err error)
	rc, _, err := ghClient.Repositories.DownloadReleaseAsset(ctx, owner, repo, *a.ID, client)

	return rc, err
}
