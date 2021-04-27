package update

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"testing"
)

/*
type HttpClient interface {
	Do(req *http.Request) (*http.Response, error)
}
*/

/*
// MockClient is the mock client
type MockClient struct {
	DoFunc func(req *http.Request) (*http.Response, error)
}
*/

/*
var (
// GetDoFunc fetches the mock client's `Do` func
//GetDoFunc func(req *http.Request) (*http.Response, error)
)
*/

/*
// Do is the mock client's `Do` func
func (m *MockClient) Do(req *http.Request) (*http.Response, error) {
	fmt.Println("do in client mock -> returns response 200")
	return m.DoFunc(req)
}
*/

// RoundTripFunc .
type RoundTripFunc func(req *http.Request) *http.Response

// RoundTrip .
func (f RoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req), nil
}

//NewTestClient returns *http.Client with Transport replaced to avoid making real calls
func NewTestClient(fn RoundTripFunc) *http.Client {
	return &http.Client{
		Transport: RoundTripFunc(fn),
	}
}

func TestCheckForUpdate(t *testing.T) {
	scenarios := []struct {
		Name           string
		CurrentVersion string
		LatestVersion  string
		LatestURL      string
		ExpectsResult  bool
	}{
		{
			Name:           "latest is newer",
			CurrentVersion: "v0.0.1",
			LatestVersion:  "v1.0.0",
			LatestURL:      "https://github.com/reliablyhq/cli/releases/latest",
			ExpectsResult:  true,
		},
		{
			Name:           "current is built from source",
			CurrentVersion: "v0.3.0-39-g98d4d79",
			LatestVersion:  "v0.3.0",
			LatestURL:      "https://github.com/reliablyhq/cli/releases/latest",
			ExpectsResult:  false,
		},
		{
			Name:           "latest is newer than version build from source",
			CurrentVersion: "v0.3.0-39-g98d4d79",
			LatestVersion:  "v0.4.0",
			LatestURL:      "https://github.com/reliablyhq/cli/releases/latest",
			ExpectsResult:  true,
		},
		{
			Name:           "latest is current",
			CurrentVersion: "v1.0.0",
			LatestVersion:  "v1.0.0",
			LatestURL:      "https://github.com/reliablyhq/cli/releases/latest",
			ExpectsResult:  false,
		},
		{
			Name:           "latest is older",
			CurrentVersion: "v0.10.0",
			LatestVersion:  "v0.9.0",
			LatestURL:      "https://github.com/reliablyhq/cli/releases/latest",
			ExpectsResult:  false,
		},
	}

	for _, s := range scenarios {
		t.Run(s.Name, func(t *testing.T) {

			client := NewTestClient(func(req *http.Request) *http.Response {
				if url := req.URL.String(); url != "https://api.github.com/repos/OWNER/REPO/releases/latest" {
					t.Errorf("Unexpected HTTP URL: %q", url)
				}

				json := fmt.Sprintf(`{
					"tag_name": "%s",
					"html_url": "%s"
				}`, s.LatestVersion, s.LatestURL)

				return &http.Response{
					StatusCode: 200,
					Body:       io.NopCloser(bytes.NewBufferString(json)),
				}
			})

			rel, err := CheckForUpdate(client, tempFilePath(), "OWNER/REPO", s.CurrentVersion)
			if err != nil {
				t.Fatal(err)
			}

			if !s.ExpectsResult {
				if rel != nil {
					t.Fatal("expected no new release")
				}
				return
			}
			if rel == nil {
				t.Fatal("expected to report new release")
			}

			if rel.Version != s.LatestVersion {
				t.Errorf("Version: %q", rel.Version)
			}
			if rel.URL != s.LatestURL {
				t.Errorf("URL: %q", rel.URL)
			}
		})
	}
}

func tempFilePath() string {
	file, err := os.CreateTemp("", "")
	if err != nil {
		log.Fatal(err)
	}
	os.Remove(file.Name())
	return file.Name()
}
