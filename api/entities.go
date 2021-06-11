package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/reliablyhq/cli/core/entities"
)

func CreateEntity(client *Client, hostname string, org string, entity entities.Entity) error {

	var version string
	version = entity.Version()
	version = strings.ToLower(version)

	var kind string
	kind = plural(entity.Kind())
	kind = strings.ToLower(kind)

	path := requestPath(version, kind, org)

	bodyBytes, err := json.Marshal(entity)
	if err != nil {
		return err
	}

	body := bytes.NewBuffer(bodyBytes)

	return client.RESTv2(hostname, http.MethodPut, path, body, nil)
}

func GetObjectiveResults(client *Client, hostname string, version string, kind string, org string) (*[]entities.ObjectiveResultResponse, error) {

	var entitiesResult *[]entities.ObjectiveResultResponse
	path := requestPath(version, kind, org)
	err := client.RESTv2(hostname, http.MethodGet, path, nil, &entitiesResult)

	if err != nil {
		return nil, fmt.Errorf("failed to make API call: %w", err)
	}

	return entitiesResult, nil

}

func requestPath(version, kind, org string) string {

	return fmt.Sprintf("%s/%s/%s/%s", "entities", version, org, kind)
}

// plural returns the puralized string,
// append trailing 's' if not already ending with it
func plural(s string) string {
	if !strings.HasSuffix(s, "s") {
		s = fmt.Sprintf("%ss", s)
	}
	return s
}
