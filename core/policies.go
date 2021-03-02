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

func policyDir(workspace, platform string, extras ...string) string {
	lplatform := strings.ToLower(platform)
	folder := filepath.Join(workspace, "policies", lplatform)
	for _, value := range extras {
		folder = fmt.Sprint(folder, "/", value)
	}
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
func FetchPolicy(workspace, platform, id string) (string, error) {
	// check whether policy is already in cache folder
	// or download it from GitHub
	// and returns its content

	var ppath = policyPath(workspace, platform, id)
	if _, err := os.Stat(ppath); os.IsNotExist(err) {
		// policy is not yet in local cache
		ppath, err = downloadPolicyToCache(workspace, platform, id)

		if err != nil {
			return "", err
		}
	}

	return ppath, nil
}

// downloadPolicyToCache downloads a given policy (by name for a targeted platform)
// into the .reliably local policies cache
func downloadPolicyToCache(workspace, platform, id string) (string, error) {
	nameParts := strings.Split(id, "/")
	var pdir string
	if len(nameParts) > 1 {
		pdir = policyDir(workspace, platform, nameParts[:len(nameParts)-1]...)
	} else {
		pdir = policyDir(workspace, platform)
	}

	ppath := policyPath(workspace, platform, id)

	lplatform := strings.ToLower(platform)
	lname := strings.ToLower(id)
	url := fmt.Sprintf(policyURL, lplatform, lname)

	_ = os.MkdirAll(pdir, 0700) // ensure to create sub-folders if not exist yet

	err := http.DownloadFile(ppath, url)
	if err != nil {
		log.Debug(fmt.Sprintf("Cannot download policy '%v/%v' from '%v'", platform, id, url))
		log.Debug(err)
		return "", err
	}

	return ppath, nil
}
