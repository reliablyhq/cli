package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/reliablyhq/cli/core"
	"github.com/reliablyhq/cli/core/report"
)

func SendReport(org, service string, r *report.Report) error {
	client := &Client{http: AuthHTTPClient(core.Hostname())}
	orgID, err := CurrentUserOrganizationID(client, core.Hostname())
	if err != nil {
		return err
	}
	path := fmt.Sprintf("orgs/%s/services/%s/reports", orgID, service)

	var body bytes.Buffer
	if err := json.NewEncoder(&body).Encode(r); err != nil {
		return fmt.Errorf("failed to serialize: %s", err)
	}

	return client.REST(core.Hostname(), http.MethodPost, path, &body, r)
}
