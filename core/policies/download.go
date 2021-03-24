package policies

import (
	"fmt"
	"strings"

	http "github.com/reliablyhq/cli/utils"
	log "github.com/sirupsen/logrus"
)

const (
	policyURL = "https://static.reliably.com/opa/%s.rego"
)

type Downloader interface {
	DownloadPolicy(string) ([]byte, error)
}

type PolicyDownloader struct {
	Downloader
}

func (pdown *PolicyDownloader) DownloadPolicy(id string) ([]byte, error) {
	url := getPolicyURL(id)

	bs, err := http.DownloadURL(url)
	if err != nil {
		log.Debugf("Cannot download policy '%s' from '%s': %s", id, url, err.Error())

		if strings.HasPrefix(err.Error(), "No file found") {
			return []byte{}, ErrPolicyNotFound
		}

		return []byte{}, err
	}
	return bs, err
}

//getPolicyURL is a helper function to return the policy URL matching the
// policy ID given
func getPolicyURL(id string) string {
	url := strings.ToLower(fmt.Sprintf(policyURL, id))
	return url
}
