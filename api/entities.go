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

func GetObjectiveResults(client *Client, hostname string, version string, org string) (*[]entities.ObjectiveResultResponse, error) {
	var entitiesResult *[]entities.ObjectiveResultResponse

	path, err := requestPath(org, version, "ObjectiveResult")
	if err != nil {
		return nil, err
	}

	if err := client.RESTv2(hostname, http.MethodGet, path, nil, &entitiesResult); err != nil {
		return nil, fmt.Errorf("failed to make API call: %w", err)
	}

	return entitiesResult, nil

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

	o := strings.ToLower(org)
	v := strings.ToLower(version)
	k := strings.ToLower(kind)
	return fmt.Sprintf("entities/%s/%s/%s", o, v, k), nil
}
