package api

import (
	"bytes"
	"encoding/json"
	"fmt"

	context "github.com/reliablyhq/cli/core/context"
)

type NewExecution struct {
	ID       string `json:"id"`
	OrgID    string `json:"org_id"`
	SourceID string `json:"source_id"`
	//ContextID string `json:"context_id"`
}

// SendExecutionContext sends the current runtime context to API
func SendExecutionContext(
	client *Client, hostname string, orgID string,
	context *context.Context) (*NewExecution, error) {

	//body := bytes.NewBufferString(`{}`)

	bodyBytes, err := json.Marshal(context)
	if err != nil {
		return nil, err
	}

	body := bytes.NewBuffer(bodyBytes)

	var response *NewExecution
	path := fmt.Sprintf("orgs/%s/executions", orgID)
	err = client.REST(hostname, "POST", path, body, &response)
	return response, err
}
