package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/reliablyhq/cli/core"
	"github.com/reliablyhq/cli/core/report"
)

func SendReport(org, service string, r *report.Report) error {
	if org == "" {
		return errors.New("org cannot be empty")
	}

	if service == "" {
		return errors.New("service cannot be empty")
	}

	if r == nil {
		return errors.New("report cannot be nil")
	}

	client := AuthHTTPClient(core.Hostname())
	u, _ := url.Parse(core.Hostname())
	u.Path = fmt.Sprintf("/api/v1/orgs/%s/services/%s/reports", org, service)

	var body bytes.Buffer
	if err := json.NewEncoder(&body).Encode(r); err != nil {
		return fmt.Errorf("failed to serialize: %s", err)
	}

	req := http.Request{
		URL:    u,
		Method: http.MethodPost,
		Body:   ioutil.NopCloser(&body),
	}

	res, err := client.Do(&req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode >= 400 {
		return fmt.Errorf("an error occured while attempting to submit the report: %s", res.Status)
	}

	return nil
}
