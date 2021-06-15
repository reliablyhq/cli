package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/reliablyhq/cli/api"
	"github.com/reliablyhq/cli/config"
	"github.com/reliablyhq/cli/core"
	"github.com/reliablyhq/cli/utils"
)

const (
	// The "GitHub CLI" OAuth app
	githubClientID = "9d58eb289eaf9ac854b2"

	// The "GitLab CLI" OAuth app
	gitlabClientID = "6056bf85b632c19d3a56617970c680ef9f6d0aca28283266d21d71abdabbcb7f"
)

type AuthProvider int

const (
	AuthWithGithub AuthProvider = iota
	AuthWithGitlab
)

func (p AuthProvider) String() string {
	switch p {
	case AuthWithGithub:
		return "github"
	case AuthWithGitlab:
		return "gitlab"
	}
	return ""
}

// AuthorizeURL returns the URL of the OAuth authorization endpoint
func (p AuthProvider) AuthorizeURL() string {
	switch p {
	case AuthWithGithub:
		return "https://github.com/login/oauth/authorize"
	case AuthWithGitlab:
		return "https://gitlab.com/oauth/authorize"
	}
	return ""
}

func (p AuthProvider) Scopes() string {
	switch p {
	case AuthWithGithub:
		return "user:email"
	case AuthWithGitlab:
		return "read_user"
	}
	return ""
}

func (p AuthProvider) ClientID() string {
	switch p {
	case AuthWithGithub:
		return githubClientID
	case AuthWithGitlab:
		return gitlabClientID
	}
	return ""
}

// localServerFlow opens the authentication page for a provider
// in a browser tab, then returns the authorization state & code
func localServerFlow(provider AuthProvider) (state string, code string, err error) {
	state, _ = utils.RandomString(20)
	code = ""

	listener, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		return
	}
	port := listener.Addr().(*net.TCPAddr).Port

	localhost := "127.0.0.1"
	callbackPath := fmt.Sprintf("/login/with/%s/authorized", provider)
	callbackURL := fmt.Sprintf("http://%s:%d%s", localhost, port, callbackPath)
	scopes := provider.Scopes()
	clientID := provider.ClientID()

	log.Debugf("Authorized callback url %s", callbackURL)

	q := url.Values{}
	q.Set("client_id", clientID)
	q.Set("redirect_uri", callbackURL)
	q.Set("scope", scopes)
	q.Set("state", state)
	q.Set("response_type", "code")

	startURL := fmt.Sprintf("%s?%s", provider.AuthorizeURL(), q.Encode())
	log.Debugf("open %s\n", startURL)
	err = utils.OpenInBrowser(startURL)
	if err != nil {
		return
	}

	_ = http.Serve(listener, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Debugf("server handler: %s\n", r.URL.Path)
		if r.URL.Path != callbackPath {
			w.WriteHeader(404)
			return
		}
		defer listener.Close()
		rq := r.URL.Query()
		if state != rq.Get("state") {
			fmt.Fprintf(w, "Error: state mismatch")
			return
		}
		log.Debugf("server received query params %s", rq)
		code = rq.Get("code")
		log.Debugf("server received code %q\n", code)

		w.Header().Add("content-type", "text/html")
		//fmt.Fprintf(w, "<p>You have successfully authenticated. You may now close this page.</p>")
		fmt.Fprint(w, oauthSuccessPage)

		/*
			if oa.WriteSuccessHTML != nil {
				oa.WriteSuccessHTML(w)
			} else {
				fmt.Fprintf(w, "<p>You have successfully authenticated. You may now close this page.</p>")
			}
		*/
	}))

	return
}

// authorizeToAPI proxies the OAuth provider callback to the API
// which is responsible for checking the OAuth access token,
// settings up a new user account, if needed, and returns a valid
// reliably access token for further authenticated API calls
func authorizeToAPI(
	hostname string, provider AuthProvider, state string, code string) (
	accessToken string, username string, err error) {

	q := url.Values{}
	// NB : we cannot send the state to API for authorization verification
	// or we end up with the following error:
	// mismatching_state: CSRF Warning! State not equal in request and response
	//q.Set("state", state)
	q.Set("code", code)

	authorizedURL := core.BaseHttpUrl(config.Hostname) + fmt.Sprintf("login/with/cli/%s/authorized", provider)
	httpClient := api.UnsecureHTTPClient(config.Hostname)
	// increase timeout for login call to API that can take a bit of extra time
	// Rather than removing the timeout, we make it very long, user will
	// problably be interrupting with ctrl-c before the end of it
	httpClient.Timeout = time.Duration(60 * time.Second)

	req, err := http.NewRequest("GET", authorizedURL, nil)
	req.URL.RawQuery = q.Encode()

	start := time.Now()
	resp, err := httpClient.Do(req)
	duration := time.Since(start).Seconds()
	log.Debugf("API call to %s took %v seconds", req.URL.Path, duration)
	if err != nil {
		log.Debug(err)
		err = errors.New("error while obtaining access token")
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("HTTP %d error while obtaining access token", resp.StatusCode)
		return
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		err = errors.New("unable to read authorized response")
		return
	}

	var response map[string]interface{}

	err = json.Unmarshal(b, &response)
	if err != nil {
		return
	}

	accessToken = response["access_token"].(string)
	username = response["username"].(string)

	return
}

func authFlow(hostname string, provider AuthProvider) (accessToken string, username string, err error) {

	// Authentication flow with provider
	state, code, err := localServerFlow(provider)
	if err != nil {
		return
	}

	// Proxying authorization code to API
	// - for oauth access token validation,
	// - retrieval of user info from OAuth provider
	// - new user account creation, with default org & api access token
	// -> returns API access token & current reliably username
	accessToken, username, err = authorizeToAPI(hostname, provider, state, code)

	return
}

const oauthSuccessPage = `
<!doctype html>
<head>
  <link rel="icon" type="image/png" href="https://reliably.com/favicon.ico" />
  <meta charset="utf-8">
  <title>Success: Reliably CLI</title>
</head>
<style type="text/css">
  body {
    color: #1B1F23;
    background: #F6F8FA;
    font-size: 14px;
    font-family: -apple-system, "Segoe UI", Helvetica, Arial, sans-serif;
    line-height: 1.5;
    max-width: 620px;
    margin: 28px auto;
    text-align: center;
  }

  .logo-mark {
    height: 5em;
    width: auto;
  }

  h1 {
    font-size: 24px;
    margin-bottom: 0;
  }

  p {
    margin-top: 0;
  }

  .box {
    border: 1px solid #E1E4E8;
    background: white;
    padding: 24px;
    margin: 28px;
  }
</style>

<body>
  <svg class="logo-mark" focusable="false" aria-hidden="true" xmlns="http://www.w3.org/2000/svg" width="24" height="24"
    viewBox="0 0 24 24">
    <circle cx="12" cy="12" r="12" fill="#d81c36" />
    <path d="M0,1.714V4H4V0H1.6A1.682,1.682,0,0,0,0,1.714Z" transform="translate(4 14)" fill="#fff" />
    <path
      d="M12.973,13a3.841,3.841,0,0,1-3.315-2.141L8.937,9.482a2.482,2.482,0,0,0-2.018-1.07H4.587A4.449,4.449,0,0,1,1.728,7.354,4.94,4.94,0,0,1,.11,4.637,5.38,5.38,0,0,1,0,3.543V0H10.09A5.977,5.977,0,0,1,13.8,1.013a3.538,3.538,0,0,1,1.333,2.963c0,1.921-1.2,3.19-3.46,3.67a2.766,2.766,0,0,1,1.009.918l.865,1.376a1.276,1.276,0,0,0,.971.459h.038a1.483,1.483,0,0,0,.722-.3L16,11.929A4.6,4.6,0,0,1,12.973,13ZM4.469,2.754V5.812H8.793A1.842,1.842,0,0,0,10.8,4.338a.38.38,0,0,0,.011-.111c-.03-.923-.784-1.474-2.018-1.474Z"
      transform="translate(4 6)" fill="#ffffff" />
  </svg>
  <div class="box">
    <h1>Successfully authenticated Reliably CLI</h1>
    <p>You may now close this tab and return to the terminal.</p>
  </div>
</body>
`
