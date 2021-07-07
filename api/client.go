package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/reliablyhq/cli/config"
	"github.com/reliablyhq/cli/core"
	"github.com/reliablyhq/cli/version"
	v "github.com/reliablyhq/cli/version"
)

// ClientOption represents an argument to NewClient
type ClientOption = func(http.RoundTripper) http.RoundTripper

// NewHTTPClient initializes an http.Client with a default timeout
func NewHTTPClient(opts ...ClientOption) *http.Client {
	tr := http.DefaultTransport
	for _, opt := range opts {
		tr = opt(tr)
	}
	return &http.Client{
		Timeout:   time.Second * 10,
		Transport: tr,
	}
}

/*
// NewHTTPClient initializes an HTTP client with a default timeout
func NewHTTPClient() *http.Client {
	return &http.Client{
		Timeout: time.Second * 2,
	}
}
*/

// Client facilitates making HTTP requests to the Reliably API
type Client struct {
	http *http.Client
}

// NewClient initializes a Client
func NewClient(opts ...ClientOption) *Client {
	client := &Client{http: NewHTTPClient(opts...)}
	return client
}

// NewClientFromHTTP takes in an http.Client instance
func NewClientFromHTTP(httpClient *http.Client) *Client {
	client := &Client{http: httpClient}
	return client
}

// AddHeader turns a RoundTripper into one that adds a request header
func AddHeader(name, value string) ClientOption {
	return func(tr http.RoundTripper) http.RoundTripper {
		return &funcTripper{roundTrip: func(req *http.Request) (*http.Response, error) {
			if req.Header.Get(name) == "" {
				req.Header.Add(name, value)
			}
			return tr.RoundTrip(req)
		}}
	}
}

// AddHeaderFunc is an AddHeader that gets the string value from a function
func AddHeaderFunc(name string, getValue func(*http.Request) (string, error)) ClientOption {
	return func(tr http.RoundTripper) http.RoundTripper {
		return &funcTripper{roundTrip: func(req *http.Request) (*http.Response, error) {
			if req.Header.Get(name) != "" {
				return tr.RoundTrip(req)
			}
			value, err := getValue(req)
			if err != nil {
				return nil, err
			}
			if value != "" {
				req.Header.Add(name, value)
			}
			return tr.RoundTrip(req)
		}}
	}
}

type funcTripper struct {
	roundTrip func(*http.Request) (*http.Response, error)
}

func (tr funcTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return tr.roundTrip(req)
}

// HTTPError is an error returned by a failed API call
type HTTPError struct {
	StatusCode int
	RequestURL *url.URL
	Message    string
}

func (err HTTPError) Error() string {
	if err.Message != "" {
		return fmt.Sprintf("HTTP %d: %s (%s)", err.StatusCode, err.Message, err.RequestURL)
	}
	return fmt.Sprintf("HTTP %d (%s)", err.StatusCode, err.RequestURL)
}

func HandleHTTPError(resp *http.Response) error {
	httpError := HTTPError{
		StatusCode: resp.StatusCode,
		RequestURL: resp.Request.URL,
	}

	if !jsonTypeRE.MatchString(resp.Header.Get("Content-Type")) {
		httpError.Message = resp.Status
		return httpError
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		httpError.Message = err.Error()
		return httpError
	}

	// TODO unmarshal the error message properly
	httpError.Message = string(body)

	return httpError
}

// REST performs a REST request and parses the response.
func (c Client) REST(hostname string, method string, p string, body io.Reader, data interface{}) error {

	url := core.RESTPrefix(hostname) + p
	log.Debugf("[api.REST] %s %s", method, url)
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("metadata-client", fmt.Sprintf("reliably-cli:%s", version.Version))

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	success := resp.StatusCode >= 200 && resp.StatusCode < 300
	if !success {
		return HandleHTTPError(resp)
	}

	if resp.StatusCode == http.StatusNoContent {
		return nil
	}

	if data != nil {
		if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
			return err
		}
	}

	return nil
}

// RESTv2 performs a REST request and parses the response.
// Compared to the previous version, this does not automatically use
// the API prefix to construct the url, only the base HTTP location of the server
func (c Client) RESTv2(hostname string, method string, p string, body io.Reader, data interface{}) error {
	url := core.BaseHttpUrl(hostname) + p
	log.Debugf("[api.REST] %s %s", method, url)
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("metadata-client", fmt.Sprintf("reliably-cli:%s", version.Version))

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	success := resp.StatusCode >= 200 && resp.StatusCode < 300
	if !success {
		return HandleHTTPError(resp)
	}

	if resp.StatusCode == http.StatusNoContent {
		return nil
	}

	if data != nil {
		if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
			return err
		}
	}

	return nil
}

var jsonTypeRE = regexp.MustCompile(`[/+]json($|;)`)

// UnsecureHTTPClient returns a non-authenticated HTTP client
func UnsecureHTTPClient(hostname string) *http.Client {
	var opts []ClientOption

	opts = append(opts,
		AddHeader("User-Agent", fmt.Sprintf("Reliably CLI v%s", v.Version)),
		AddHeader("Accept", "application/json"),
	)

	return NewHTTPClient(opts...)
}

// AuthHTTPClient returns an authenticated HTTP client
func AuthHTTPClient(hostname string) *http.Client {
	var opts []ClientOption
	var token string

	opts = append(opts,
		AddHeader("User-Agent", fmt.Sprintf("Reliably CLI v%s", v.Version)),
		AddHeader("Accept", "application/json"),
	)

	token = config.GetTokenFor(hostname)

	if token != "" {
		opts = append(opts,
			AddHeader("Authorization", fmt.Sprintf("Bearer %s", token)),
		)
	}

	return NewHTTPClient(opts...)
}

/*

// generic authenticated HTTP client for commands
func NewHTTPClient(cfg config.Config, appVersion string, setAccept bool) *http.Client {
	var opts []api.ClientOption

	opts = append(opts,
		api.AddHeader("User-Agent", fmt.Sprintf("GitHub CLI %s", appVersion)),
		api.AddHeaderFunc("Authorization", func(req *http.Request) (string, error) {
			hostname := req.URL.Hostname()
			if token, err := cfg.Get(hostname, "oauth_token"); err == nil && token != "" {
				return fmt.Sprintf("token %s", token), nil
			}
			return "", nil
		}),
	)

	if setAccept {
		opts = append(opts, api.AddHeader("Accept", "application/json"))
	}

	return api.NewHTTPClient(opts...)
}
*/
