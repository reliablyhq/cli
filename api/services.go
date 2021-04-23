package api

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/reliablyhq/cli/core/manifest"
)

// PushManifest - records the manifest via the API backend.
func PushManifest(client *Client, hostname, orgID string, m *manifest.Manifest) error {
	// path :=

	var body bytes.Buffer
	if err := json.NewEncoder(&body).Encode(m); err != nil {
		return fmt.Errorf("failed to serialize: %s", err)
	}

	path := fmt.Sprintf("orgs/%s/services", orgID)
	var response interface{}
	if err := client.REST(hostname, "POST",
		path, &body, response); err != nil {
		return err
	}

	return nil
}

// PullManifest - downloads current manifest
func PullManifest(client *Client, hostname, orgID string) (*manifest.Manifest, error) {
	var m manifest.Manifest

	path := fmt.Sprintf("orgs/%s/services", orgID)
	if err := client.REST(hostname, "GET",
		path, nil, &m); err != nil {
		return nil, err
	}

	return &m, nil
}
