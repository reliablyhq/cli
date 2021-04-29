package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/reliablyhq/cli/core"
	"github.com/reliablyhq/cli/core/manifest"
	log "github.com/sirupsen/logrus"
)

var apiURL = &url.URL{Scheme: "https", Host: core.Hostname()}

// PushServiceManifest - records the manifest via the API backend for a given service
// note that this will append the service to the main organisational manifest.
func PushServiceManifest(service string, m *manifest.Manifest) error {
	if service == "" {
		return errors.New("service cannot be empty")
	}

	client := AuthHTTPClient(apiURL.Host)
	orgID, err := CurrentUserOrganizationID(&Client{http: client}, apiURL.Host)
	if err != nil {
		return err
	}

	var body bytes.Buffer
	if err := json.NewEncoder(&body).Encode(m); err != nil {
		return fmt.Errorf("failed to serialize: %s", err)
	}

	u, _ := url.Parse(apiURL.String())
	u.Path = fmt.Sprintf("/api/v1/orgs/%s/services/%s", orgID, service)

	req := http.Request{
		URL:    u,
		Method: http.MethodPut,
		Body:   ioutil.NopCloser(&body),
	}

	log.Debugf("%s %s", req.Method, u)
	res, err := client.Do(&req)
	if err != nil {
		return err
	}

	if res.StatusCode != 200 {
		return fmt.Errorf("its dead, Jim! %s", res.Status)
	}

	return nil
}

// PushManifest - push entire manifest
func PushManifest(m *manifest.Manifest) error {
	client := AuthHTTPClient(apiURL.Host)
	orgID, err := CurrentUserOrganizationID(&Client{http: client}, apiURL.Host)
	if err != nil {
		return err
	}

	var body bytes.Buffer
	if err := json.NewEncoder(&body).Encode(m); err != nil {
		return fmt.Errorf("failed to serialize: %s", err)
	}

	u, _ := url.Parse(apiURL.String())
	u.Path = fmt.Sprintf("/api/v1/orgs/%s/services", orgID)

	req := http.Request{
		URL:    u,
		Method: http.MethodPut,
		Body:   ioutil.NopCloser(&body),
	}

	log.Debugf("%s %s", req.Method, u)
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
func PullServiceManifest(service string) (*manifest.Manifest, error) {
	if service == "" {
		return nil, errors.New("service cannot be empty")
	}

	client := AuthHTTPClient(apiURL.Host)

	orgID, err := CurrentUserOrganizationID(&Client{http: client}, apiURL.Host)
	if err != nil {
		return nil, err
	}

	u, _ := url.Parse(apiURL.String())
	u.Path = fmt.Sprintf("/api/v1/orgs/%s/services/%s", orgID, service) // get all by default
	// u.Path = fmt.Sprintf("/api/v1/orgs/%s/services/%s", orgID, service)

	req := http.Request{
		URL:    u,
		Method: http.MethodGet,
	}

	log.Debugf("%s %s", req.Method, u)
	res, err := client.Do(&req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode == 404 {
		return nil, nil
	}

	if res.StatusCode != 200 {
		return nil, fmt.Errorf("bad response from server %s", res.Status)
	}

	var s manifest.Service
	if err := json.NewDecoder(res.Body).Decode(&s); err != nil {
		return nil, fmt.Errorf("failed to deserialize: %s", err)
	}

	var m manifest.Manifest
	m.Services = append(m.Services, &s)
	return &m, nil
}

// PullServiceManifest - downloads current manifest
func PullManifest() (*manifest.Manifest, error) {
	client := AuthHTTPClient(apiURL.Host)
	orgID, err := CurrentUserOrganizationID(&Client{http: client}, apiURL.Host)
	if err != nil {
		return nil, err
	}

	u, _ := url.Parse(apiURL.String())
	u.Path = fmt.Sprintf("/api/v1/orgs/%s/services", orgID) // get all by default
	// u.Path = fmt.Sprintf("/api/v1/orgs/%s/services/%s", orgID, service)

	req := http.Request{
		URL:    u,
		Method: http.MethodGet,
	}

	log.Debugf("%s %s", req.Method, u)
	res, err := client.Do(&req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode == 404 {
		return nil, nil
	}

	if res.StatusCode != 200 {
		return nil, fmt.Errorf("bad response from server %s", res.Status)
	}

	var m manifest.Manifest
	if err := json.NewDecoder(res.Body).Decode(&m); err != nil {
		return nil, fmt.Errorf("failed to deserialize: %s", err)
	}

	return &m, nil
}

func ServiceExists(_, service string) (bool, error) {
	// if org == "" {
	// 	return false, errors.New("org cannot be empty")
	// }

	if service == "" {
		return false, errors.New("service cannot be empty")
	}

	client := AuthHTTPClient(apiURL.Host)

	orgID, err := CurrentUserOrganizationID(NewClient(), apiURL.Host)
	if err != nil {
		return false, err
	}

	u, _ := url.Parse(apiURL.String())
	u.Path = fmt.Sprintf("/api/v1/orgs/%s/services/%s", orgID, service)

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
