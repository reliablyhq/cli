package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/reliablyhq/cli/core/entities"
)

func CreateEntity(client *Client, hostname string, org string, entity entities.Entity) error {
	path, err := requestPath(org, entity.Version(), entity.Kind())
	if err != nil {
		return err
	}

	var body bytes.Buffer
	if err := json.NewEncoder(&body).Encode(entity); err != nil {
		return err
	}

	return client.RESTv2(hostname, http.MethodPut, path, &body, nil)
}

func Query(client *Client, hostname string, version string, org string, query QueryBody) (*QueryResponse, error) {

	var response QueryResponse

	path, err := requestPath(org, version, "query")
	if err != nil {
		return nil, err
	}

	bodyBytes, err := json.Marshal(query)
	fmt.Printf("%s", bodyBytes)
	if err != nil {
		return nil, err
	}

	body := bytes.NewBuffer(bodyBytes)

	if err := client.RESTv2(hostname, http.MethodPost, path, body, &response); err != nil {
		return nil, fmt.Errorf("failed to make API call: %w", err)
	}

	return &response, nil

}

func requestPath(org, version, kind string) (string, error) {
	if org == "" {
		return "", errors.New("org is empty")
	}

	if version == "" {
		return "", errors.New("version is empty")
	}

	if kind == "" {
		return "", errors.New("kind is empty")
	}

	path := fmt.Sprintf("entities/%s/%s/%s", org, version, kind)

	return strings.ToLower(path), nil
}
