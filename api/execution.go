package api

import (
	"bytes"
	"encoding/json"
	"fmt"

	context "github.com/reliablyhq/cli/core/context"
)

// SendExecutionContext sends the current runtime context to API
func SendExecutionContext(
	client *Client, hostname string, orgID string,
	context *context.Context) (string, error) {

	//body := bytes.NewBufferString(`{}`)

	bodyBytes, err := json.Marshal(context)
	if err != nil {
		return "", err
	}

	body := bytes.NewBuffer(bodyBytes)

	var response struct {
		ID string `json:"id"`
	}

	// DEBUG
	fmt.Println(string(bodyBytes))

	path := fmt.Sprintf("orgs/%s/executions", orgID)
	err = client.REST(hostname, "POST", path, body, &response)
	return response.ID, err
}
