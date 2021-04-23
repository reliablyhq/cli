package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"

	"github.com/reliablyhq/cli/core/manifest"
	"github.com/reliablyhq/cli/version"
)

var hostname = "https://api.reliably.com"

func init() {
	if version.IsDevVersion() {
		hostname = "https://api.reliably.dev"
	}

	if x := os.Getenv("RELIABLY_HOST"); x != "" {
		hostname = x
	}
}

// PushServiceManifest - records the manifest via the API backend.
func PushServiceManifest(org, service string, m *manifest.Manifest) error {
	client := AuthHTTPClient(hostname)

	var body bytes.Buffer
	if err := json.NewEncoder(&body).Encode(m); err != nil {
		return fmt.Errorf("failed to serialize: %s", err)
	}

	u, _ := url.Parse(hostname)
	u.Path = fmt.Sprintf("/api/v1/orgs/%s/services/%s", org, service)

	req := http.Request{
		URL:    u,
		Method: http.MethodPut,
		Body:   ioutil.NopCloser(&body),
	}

	res, err := client.Do(&req)
	if err != nil {
		return err
	}

	if res.StatusCode != 200 {
		return fmt.Errorf("its dead, Jim! %s", res.Status)
	}

	return nil
}

// PullServiceManifest - downloads current manifest
func PullServiceManifest(org, service string) (*manifest.Manifest, error) {
	client := AuthHTTPClient(hostname)

	u, _ := url.Parse(hostname)
	u.Path = fmt.Sprintf("/api/v1/orgs/%s/services/%s", org, service)

	req := http.Request{
		URL:    u,
		Method: http.MethodGet,
	}

	res, err := client.Do(&req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return nil, fmt.Errorf("its dead, Jim! %s", res.Status)
	}

	var m manifest.Manifest
	if err := json.NewDecoder(res.Body).Decode(&m); err != nil {
		return nil, fmt.Errorf("failed to deserialize: %s", err)
	}

	return &m, nil
}

func ServiceExists(org, service string) (bool, error) {
	client := AuthHTTPClient(hostname)

	u, _ := url.Parse(hostname)
	u.Path = fmt.Sprintf("/api/v1/orgs/%s/services/%s", org, service)

	req := http.Request{
		URL:    u,
		Method: http.MethodGet,
	}

	res, err := client.Do(&req)
	if err != nil {
		return false, err
	}
	defer res.Body.Close()

	if res.StatusCode >= 200 && res.StatusCode < 400 {
		return true, nil
	}

	if res.StatusCode == 404 {
		return false, nil
	}

	return false, fmt.Errorf("an error occured while retrieveing the serivce: %s", res.Status)
}
