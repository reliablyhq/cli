package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/url"
	"time"

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

type Execution struct {
	ID          string       `json:"id"`
	Date        time.Time    `json:"created_on"`
	Suggestions []Suggestion `json:"suggestions"`
}

type Suggestion struct {
	ID   string           `json:"id"`
	Date time.Time        `json:"created_on"`
	Data *core.Suggestion `json:"data"`
}

type Page struct {
	Cursor      string `json:"cursor"`
	HasNextPage bool   `json:"has_next_page"`
}

type SuggestionHistory struct {
	PageInfo   Page        `json:"page_info"`
	Executions []Execution `json:"executions"`
}

func GetSuggestionHistory(client *Client, hostname string,
	orgID string, sourceID string, cursor string) (*SuggestionHistory, error) {

	var history *SuggestionHistory

	path := fmt.Sprintf("orgs/%s/suggestions", orgID)
	if cursor != "" {
		params := url.Values{}
		params.Add("cursor", cursor)
		params.Add("source_id", sourceID)
		path = fmt.Sprintf("%s?%s", path, params.Encode())
	}

	err := client.REST(hostname, "GET", path, nil, &history)

	return history, err
}
