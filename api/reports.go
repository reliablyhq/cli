package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	log "github.com/sirupsen/logrus"

	"github.com/reliablyhq/cli/core"
	"github.com/reliablyhq/cli/core/report"
)

func SendReport(client *Client, orgID string, r *report.Report) (string, error) {
	path := fmt.Sprintf("orgs/%s/reports", orgID)

	var body bytes.Buffer
	if err := json.NewEncoder(&body).Encode(r); err != nil {
		return "", fmt.Errorf("failed to serialize: %s", err)
	}

	type savedReport struct {
		ID string `json:"id"`
	}
	var response *savedReport

	err := client.REST(core.Hostname(), http.MethodPost, path, &body, &response)
	if err == nil {
		log.Debugf("Report %s has been saved", response.ID)
	}

	return response.ID, err
}
