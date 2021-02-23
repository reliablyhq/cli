package api

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/reliablyhq/cli/core"
)

// RecordSuggestions records on the API the suggestions for a given execution
// suggestions can be recorded by batch or one at a time (maybe not ideal)
// and will be appended on the backend side
func RecordSuggestions(client *Client, hostname string,
	orgID string, execID string, suggestions *[]*core.Suggestion) error {

	/*
		var data map[string]interface{} = map[string]interface{}{
			"suggestions": suggestions,
		}
	*/

	bodyBytes, err := json.Marshal(suggestions)
	if err != nil {
		return fmt.Errorf("failed to serialize: %w", err)
	}

	body := bytes.NewBuffer(bodyBytes)

	path := fmt.Sprintf("orgs/%s/executions/%s/suggestions", orgID, execID)
	err = client.REST(hostname, "POST", path, body, nil)
	if err != nil {
		return fmt.Errorf("failed to make API call: %w", err)
	}

	return nil
}
