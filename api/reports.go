package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/reliablyhq/cli/core"
	"github.com/reliablyhq/cli/core/report"
)

func SendReport(client *Client, orgID string, r *report.Report) error {
	path := fmt.Sprintf("orgs/%s/reports", orgID)

	var body bytes.Buffer
	if err := json.NewEncoder(&body).Encode(r); err != nil {
		return fmt.Errorf("failed to serialize: %s", err)
	}

	return client.REST(core.Hostname(), http.MethodPost, path, &body, r)
}
