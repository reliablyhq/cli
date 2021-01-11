package auth

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"

	//"time"

	log "github.com/sirupsen/logrus"

	"github.com/reliablyhq/cli/api"
	"github.com/reliablyhq/cli/utils"
)

const (
	// The "GitHub CLI" OAuth app
	githubClientID = "9d58eb289eaf9ac854b2"
	// This value is safe to be embedded in version control
	githubClientSecret = "aaee7c6566f8e88458497d3f124f9f409bdf35cf"
)

type AuthMode int

const (
	authWithGithub AuthMode = iota
	authWithGitlab
)

type OAuthProfile struct {
	Sub      string `json:"sub"`
	Email    string `json:"email"`
	Fullname string `json:"fullname"`
	Username string `json:"username"`
	//URL string `json:"url"`
	//Picture string `json:"picture"`
}

type OAuthToken struct {
	Token string `json:"access_token"`
	Type  string `json:"token_type"`
	Scope string `json:"scope"`
}

type AuthorizedOAuth struct {
	Provider string       `json:"provider"`
	Token    OAuthToken   `json:"token"`
	Profile  OAuthProfile `json:"profile"`
}

// UserInfo represents the OpenID UserInfo response
type UserInfo struct {
	Sub               string `json:"sub"`
	Name              string `json:"name"`
	Email             string `json:"email"`
	PreferredUsername string `json:"preferred_username"`
	Profile           string `json:"profile"`
	Picture           string `json:"picture"`
	Website           string `json:"website"`
}

type GithubOAuthProfile struct {
	Sub               int    `json:"id"`
	Name              string `json:"name"`
	Email             string `json:"email"`
	PreferredUsername string `json:"login"`
	//Profile string `json:"html_url"`
	//Picture string `json:"avatar_url"`
	//Website string `json:"blog"`

}

func localServerFlow() (accessToken *OAuthToken, err error) {
	state, _ := utils.RandomString(20)

	code := ""
	listener, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		return
	}
	port := listener.Addr().(*net.TCPAddr).Port

	localhost := "127.0.0.1"
	callbackPath := "/login/with/github/authorized"
	scopes := "user:email"

	q := url.Values{}
	q.Set("client_id", githubClientID)
	q.Set("redirect_uri", fmt.Sprintf("http://%s:%d%s", localhost, port, callbackPath))
	q.Set("scope", scopes)
	q.Set("state", state)

	startURL := fmt.Sprintf("https://github.com/login/oauth/authorize?%s", q.Encode())
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
		code = rq.Get("code")
		log.Debugf("server received code %q\n", code)
		w.Header().Add("content-type", "text/html")
		//fmt.Fprintf(w, "<p>You have successfully authenticated. You may now close this page.</p>")
		fmt.Fprintf(w, oauthSuccessPage)
		/*
			if oa.WriteSuccessHTML != nil {
				oa.WriteSuccessHTML(w)
			} else {
				fmt.Fprintf(w, "<p>You have successfully authenticated. You may now close this page.</p>")
			}
		*/
	}))

	/*
		httpClient := &http.Client{
			Timeout: time.Second * 2,
		}
	*/
	httpClient := http.DefaultClient

	tokenURL := fmt.Sprintf("https://github.com/login/oauth/access_token")
	log.Debugf("POST %s\n", tokenURL)
	resp, err := httpClient.PostForm(tokenURL,
		url.Values{
			"client_id":     {githubClientID},
			"client_secret": {githubClientSecret},
			"code":          {code},
			"state":         {state},
		})
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("HTTP %d error while obtaining OAuth access token", resp.StatusCode)
		return
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	tokenValues, err := url.ParseQuery(string(body))
	if err != nil {
		return
	}

	if tokenValues.Get("access_token") == "" {
		err = errors.New("the access token could not be read from HTTP response")
	}

	accessToken = &OAuthToken{
		Token: tokenValues.Get("access_token"),
		Type:  tokenValues.Get("token_type"),
		Scope: tokenValues.Get("scope"),
	}

	return
}

func authFlow(hostname string) (string, interface{}, error) {

	// Authentication flow with provider
	oauthToken, err := localServerFlow()
	if err != nil {
		return "", "", err
	}

	// Fetch User Profile from provider once authenticated
	profile, err := getUserProfile(oauthToken.Token)
	if err != nil {
		return "", "", err
	}

	provider := "github"
	token, username, err := registerAuthorizedUser(hostname, provider, oauthToken, profile)
	if err != nil {
		return "", "", err
	}

	return token, username, nil
}

func registerAuthorizedUser(hostname string, provider string, token *OAuthToken, profile *OAuthProfile) (string, string, error) {

	// register authorized user to Reliably & fetches its access token
	var body AuthorizedOAuth = AuthorizedOAuth{
		Provider: provider,
		Token:    *token,
		Profile:  *profile,
	}

	requestByte, err := json.Marshal(body)
	if err != nil {
		return "", "", err
	}
	requestBody := bytes.NewReader(requestByte)

	apiClient := api.NewClientFromHTTP(api.UnsecureHTTPClient(hostname))
	data, err := api.LoginWithCliAuthorized(apiClient, hostname, requestBody)
	if err != nil {
		return "", "", err
	}

	return data["access_token"].(string), data["username"].(string), nil
}

func getUserProfile(token string) (profile *OAuthProfile, err error) {

	log.Debug("Fetch user profile with GH API")

	// authenticated client for GitHub
	httpClient := api.NewHTTPClient(api.AddHeader("Authorization", fmt.Sprintf("token %s", token)))

	resp, err := httpClient.Get("https://api.github.com/user")
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("HTTP %d error while obtaining GitHub user profile", resp.StatusCode)
		return
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	profileStr := string(body)

	var p GithubOAuthProfile

	err = json.Unmarshal([]byte(profileStr), &p)
	if err != nil {
		return
	}

	profile = &OAuthProfile{
		Sub:      fmt.Sprint(p.Sub),
		Email:    p.Email,
		Fullname: p.Name,
		Username: p.PreferredUsername,
	}

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
