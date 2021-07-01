package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/reliablyhq/cli/core/entities"
	"github.com/reliablyhq/cli/utils"
)

func CreateEntity(client *Client, hostname string, org string, entity entities.Entity) error {
	path := requestPath(org, entity.Version(), entity.Kind())

	var body bytes.Buffer
	if err := json.NewEncoder(&body).Encode(entity); err != nil {
		return err
	}

	return client.RESTv2(hostname, http.MethodPut, path, &body, nil)
}

func GetObjectiveResults(client *Client, hostname string, version string, org string) (*[]entities.ObjectiveResultResponse, error) {

	var entitiesResult *[]entities.ObjectiveResultResponse

	shortVersion, ok := utils.GetShortVersion(version)
	if !ok {
		return nil, fmt.Errorf("version %v not supported", version)
	}

	path := requestPath(shortVersion, "objective-results", org)
	err := client.RESTv2(hostname, http.MethodGet, path, nil, &entitiesResult)

	if err != nil {
		return nil, fmt.Errorf("failed to make API call: %w", err)
	}

	return entitiesResult, nil

}

func requestPath(org, version, kind string) string {
	o := strings.ToLower(org)
	v := strings.ToLower(version)
	k := strings.ToLower(kind)
	return fmt.Sprintf("entities/%s/%s/%s", o, v, k)
}
