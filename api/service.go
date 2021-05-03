package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/reliablyhq/cli/core"
	"github.com/reliablyhq/cli/core/manifest"
)

// PushServiceManifest - records the manifest via the API backend for a given service
// note that this will append the service to the main organisational manifest.
func PushServiceManifest(service string, m *manifest.Manifest) error {
	if service == "" {
		return errors.New("service cannot be empty")
	}

	client := &Client{http: AuthHTTPClient(core.Hostname())}
	orgID, err := CurrentUserOrganizationID(client, core.Hostname())
	if err != nil {
		return err
	}

	var body bytes.Buffer
	if err := json.NewEncoder(&body).Encode(m); err != nil {
		return fmt.Errorf("failed to serialize: %s", err)
	}

	path := fmt.Sprintf("orgs/%s/services/%s", orgID, service)
	return client.REST(core.Hostname(), http.MethodPut, path, &body, nil)
}

// PushManifest - push entire manifest
func PushManifest(m *manifest.Manifest) error {
	client := &Client{http: AuthHTTPClient(core.Hostname())}
	orgID, err := CurrentUserOrganizationID(client, core.Hostname())
	if err != nil {
		return err
	}

	var body bytes.Buffer
	if err := json.NewEncoder(&body).Encode(m); err != nil {
		return fmt.Errorf("failed to serialize: %s", err)
	}

	path := fmt.Sprintf("orgs/%s/services", orgID)
	return client.REST(core.Hostname(), http.MethodPut, path, &body, nil)
}

// PullServiceManifest - downloads current manifest
func PullServiceManifest(service string) (*manifest.Manifest, error) {
	if service == "" {
		return nil, errors.New("service cannot be empty")
	}

	client := &Client{http: AuthHTTPClient(core.Hostname())}
	orgID, err := CurrentUserOrganizationID(client, core.Hostname())
	if err != nil {
		return nil, err
	}

	var s manifest.Service
	path := fmt.Sprintf("orgs/%s/services/%s", orgID, service) // get all by default
	if err := client.REST(core.Hostname(), http.MethodGet, path, nil, &s); err != nil {
		return nil, err
	}

	var m manifest.Manifest
	m.Services = append(m.Services, &s)
	return &m, nil
}

// PullServiceManifest - downloads current manifest
func PullManifest() (*manifest.Manifest, error) {
	client := &Client{http: AuthHTTPClient(core.Hostname())}
	orgID, err := CurrentUserOrganizationID(client, core.Hostname())
	if err != nil {
		return nil, err
	}

	path := fmt.Sprintf("orgs/%s/services", orgID) // get all by default
	var m manifest.Manifest
	return &m, client.REST(core.Hostname(), http.MethodGet, path, nil, &m)
}
