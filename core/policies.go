package core

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	http "github.com/reliablyhq/cli/utils"
	log "github.com/sirupsen/logrus"
)

const (
	policyURL = "https://static.reliably.com/opa/%s/%s.rego"
)

func policyDir(workspace string, platform string) string {
	lplatform := strings.ToLower(platform)
	folder := filepath.Join(workspace, "policies", lplatform)
	return folder
}

func policyPath(workspace string, platform string, name string) string {
	pdir := policyDir(workspace, platform)
	lname := strings.ToLower(name)
	ppath := filepath.Join(pdir, fmt.Sprintf("%v.rego", lname))
	return ppath
}

// FetchPolicy ensure the policy is available in cache and returns
// its file path. If policy is not in the cache, it downloads it from GitHub
func FetchPolicy(workspace string, platform string, name string) (string, error) {
	// check whether policy is already in cache folder
	// or download it from GitHub
	// and returns its content

	var ppath = policyPath(workspace, platform, name)
	if _, err := os.Stat(ppath); os.IsNotExist(err) {
		// policy is not yet in local cache
		ppath, err = DownloadPolicyToCache(workspace, platform, name)

		if err != nil {
			return "", err
		}
	}

	return ppath, nil
}

// DownloadPolicyToCache downloads a given policy (by name for a targeted platform)
// into the .reliably local policies cache
func DownloadPolicyToCache(workspace string, platform string, name string) (string, error) {
	log.Warning("This function 'DownloadPolicyToCache' is deprecated. You shall fetch Patterns from API instead.")

	pdir := policyDir(workspace, platform)
	ppath := policyPath(workspace, platform, name)

	lplatform := strings.ToLower(platform)
	lname := strings.ToLower(name)
	url := fmt.Sprintf(policyURL, lplatform, lname)

	_ = os.MkdirAll(pdir, 0700) // ensure to create sub-folders if not exist yet

	err := http.DownloadFile(ppath, url)
	if err != nil {
		log.Debug(fmt.Sprintf("Cannot download policy '%v/%v' from '%v'", platform, name, url))
		log.Debug(err)
		return "", err
	}

	return ppath, nil
}
