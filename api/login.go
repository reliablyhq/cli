package api

import (
	//"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/reliablyhq/cli/core"
)

// LoginWithCliAuthorized pushes the oauth authorized data to the API
// for account & API access token creation
// This API shall be done with an un-authenticated http client,
// to an unsecure endpoint (that is not under base api prefix)
func LoginWithCliAuthorized(client *Client, hostname string, body io.Reader) (map[string]interface{}, error) {

	url := core.BaseHttpUrl(hostname) + "login/with/cli/authorized"

	/*
		requestByte, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		requestBody := bytes.NewReader(requestByte)
	*/

	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json; charset=utf-8")

	resp, err := client.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	success := resp.StatusCode >= 200 && resp.StatusCode < 300
	if !success {
		return nil, HandleHTTPError(resp)
	}

	if resp.StatusCode == http.StatusNoContent {
		return nil, nil
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var data map[string]interface{}

	err = json.Unmarshal(b, &data)
	if err != nil {
		return nil, err
	}

	return data, nil

}
