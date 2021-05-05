package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

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

func GetReports(client *Client, hostname string, orgID string, limit int) ([]report.Report, error) {

	type response struct {
		Reports  []report.Report `json:"reports"`
		PageInfo struct {
			Cursor      string `json:"cursor"`
			HasNextPage bool   `json:"has_next_page"`
		} `json:"page_info"`
	}

	var reports []report.Report

	var cursor string
	pageLimit := min(limit, 100) // at most, 100 results per api call

loop:

	for {

		var resp response

		path := fmt.Sprintf("orgs/%s/reports/history", orgID)
		params := url.Values{}
		params.Add("limit", fmt.Sprint(pageLimit))
		if cursor != "" {
			params.Add("cursor", cursor)
		}
		path = fmt.Sprintf("%s?%s", path, params.Encode())

		err := client.REST(hostname, http.MethodGet, path, nil, &resp)
		if err != nil {
			return nil, err
		}

		for _, r := range resp.Reports {
			reports = append(reports, r)
			if len(reports) == limit {
				break loop
			}
		}

		if !resp.PageInfo.HasNextPage {
			break
		}

		cursor = resp.PageInfo.Cursor
		pageLimit = min(pageLimit, limit-len(reports))
	}

	return reports, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
