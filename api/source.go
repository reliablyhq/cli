package api

import (
	"bytes"
	"encoding/json"
	"fmt"
)

type Source struct {
	ID   string                 `json:"id"`
	Type string                 `json:"type"`
	Meta map[string]interface{} `json:"meta"`
}

// CurrentSourceID looks up into the API/DB for a registered Source
// from the internally computed Source hash (by the CLI)
func CurrentSourceID(
	client *Client, hostname string,
	orgID string, hash string) (string, error) {

	var src Source

	var payload map[string]interface{} = map[string]interface{}{
		"hash": hash,
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to serialize: %w", err)
	}

	body := bytes.NewBuffer(bodyBytes)

	url := fmt.Sprintf("orgs/%s/sources/lookup", orgID)
	err = client.REST(hostname, "POST", url, body, &src)
	if err != nil {
		return "", err
	}

	return src.ID, nil
}
