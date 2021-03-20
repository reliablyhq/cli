package policies

import (
	"errors"
)

// ErrPolicyNotFound is returned when a policy could not be found
var ErrPolicyNotFound = errors.New("Policy not found")

/*
func policyDir(workspace, platform string, extras ...string) string {
	lplatform := strings.ToLower(platform)
	folder := filepath.Join(workspace, "policies", lplatform)
	for _, value := range extras {
		folder = fmt.Sprint(folder, "/", value)
	}
	return folder
}
*/

/*
func policyPath(workspace string, platform string, name string) string {
	pdir := policyDir(workspace, platform)
	lname := strings.ToLower(name)
	ppath := filepath.Join(pdir, fmt.Sprintf("%v.rego", lname))
	return ppath
}
*/

/*
// downloadPolicyToCache downloads a given policy (by name for a targeted platform)
// into the .reliably local policies cache
func downloadPolicyToCache(workspace, platform, path string) (string, error) {
	pathParts := strings.Split(path, "/")
	var pdir string
	if len(pathParts) > 1 {
		pdir = policyDir(workspace, platform, pathParts[:len(pathParts)-1]...)
	} else {
		pdir = policyDir(workspace, platform)
	}

	ppath := policyPath(workspace, platform, path)

	lplatform := strings.ToLower(platform)
	lname := strings.ToLower(path)
	url := fmt.Sprintf(policyURL, lplatform, lname)

	_ = os.MkdirAll(pdir, 0700) // ensure to create sub-folders if not exist yet

	err := http.DownloadFile(ppath, url)
	if err != nil {
		log.Debug(err)

		if strings.HasPrefix(err.Error(), "No file found") {
			return "", ErrPolicyNotFound
		}

		log.Debug(fmt.Sprintf("Cannot download policy '%v/%v' from '%v'", platform, path, url))
		return "", err
	}

	return ppath, nil
}
*/
