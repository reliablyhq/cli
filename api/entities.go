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

	path := fmt.Sprintf("%s/%s/%s/%s", "entities", version, org, kind)

	bodyBytes, err := json.Marshal(entity)
	if err != nil {
		return err
	}

	body := bytes.NewBuffer(bodyBytes)

	return client.RESTv2(hostname, http.MethodPut, path, body, nil)
}

// plural returns the puralized string,
// append trailing 's' if not already ending with it
func plural(s string) string {
	if !strings.HasSuffix(s, "s") {
		s = fmt.Sprintf("%ss", s)
	}
	return s
}
