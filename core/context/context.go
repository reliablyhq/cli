package context

import (
	"encoding/json"
	"fmt"
	"os"
	"os/user"
	"runtime"
	"time"

	"github.com/reliablyhq/cli/utils"
)

// StringMap represents a map of string keys to string values
type StringMap map[string]string

// ContextMap ...
type ContextMap map[string]interface{}

// Context contains all execution data useful to Reliably when running the CLI
type Context struct {
	IsCI    bool            `json:"ci"`
	Runtime *RuntimeContext `json:"runtime"`
	Environ StringMap       `json:"environ"`
	Context ContextMap      `json:"context"`
	Source  interface{}     `json:"source"` // can be one of Local, Github, Gitlab sources
}

// RuntimeContext holds the metadata of the environment running the CLI
type RuntimeContext struct {
	Datetime   time.Time `json:"datetime"`
	Username   string    `json:"username"`
	Hostname   string    `json:"hostname"`
	WorkingDir string    `json:"workingdir"`
	Command    []string  `json:"command"`
	OS         string    `json:"os"`
	Arch       string    `json:"arch"`
}

// NewRuntimeContext is a constructor to initialize a CLI execution context
func NewRuntimeContext() *RuntimeContext {

	var (
		username string
		wd       string
		hostname string
	)

	dt := time.Now()

	if cwd, err := os.Getwd(); err == nil {
		wd = cwd
	}

	if user, err := user.Current(); err == nil {
		username = user.Username
	}

	if host, err := os.Hostname(); err == nil {
		hostname = host
	}

	return &RuntimeContext{
		Datetime:   dt,
		Username:   username,
		Hostname:   hostname,
		WorkingDir: wd,
		Command:    os.Args,
		OS:         runtime.GOOS,
		Arch:       runtime.GOARCH,
	}
}

// GithubEnv is a map of targeted GitHub env variables set
// when running the CLI via a Github workflow
type GithubEnv StringMap

// GitlabEnv is a map of targeted Gitlab env variables set
// when running the CLI via a Gitlab workflow
type GitlabEnv StringMap

// GetGithubEnv returns a map of GitHub env variables
func GetGithubEnv() GithubEnv {

	keys := []string{"CI", "RUNNER_OS"}

	// Github env vars are started with the GITHUB_prefix
	ghKeys := []string{
		"ACTION", "ACTIONS", "ACTION_REF", "ACTION_REPOSITORY", "ACTOR",
		"JOB", "REF", "REPOSITORY", "REPOSITORY_OWNER", "RUN_ID", "SERVER_URL",
		"SHA", "WORKFLOW", "RUN_NUMBER",
	}

	for _, k := range ghKeys {
		key := "GITHUB_" + k
		keys = append(keys, key)
	}

	return getOsEnvs(keys)
}

// toContextMap converts a GitHub env map to ContextMap object
// This is due to map type casting
/*
func (env GithubEnv) toContextMap() *ContextMap {
	cm := make(ContextMap, len(env))
	for k, v := range env {
		cm[k] = v
	}
	return &cm
}
*/

// GetGitlabEnv returns a map of Gitlab env variables
func GetGitlabEnv() GitlabEnv {

	keys := []string{"CI", "GITLAB_CI", "GITLAB_USER_LOGIN", "HOSTNAME"} // "GITLAB_USER_NAME", "GITLAB_USER_ID", "GITLAB_USER_EMAIL",

	ciKeys := []string{
		"CI_SERVER_URL", "CI_RUNNER_EXECUTABLE_ARCH",

		"CI_COMMIT_SHORT_SHA", "CI_COMMIT_SHA", "CI_COMMIT_BRANCH",
		"CI_COMMIT_REF_NAME", "CI_COMMIT_REF_SLUG", "CI_COMMIT_TIMESTAMP",
		"CI_COMMIT_TITLE", "CI_COMMIT_MESSAGE",

		"CI_PROJECT_ID", "CI_PROJECT_NAMESPACE", "CI_PROJECT_TITLE",
		"CI_PROJECT_VISIBILITY", "CI_PROJECT_NAME", "CI_PROJECT_URL",

		"CI_PIPELINE_ID", "CI_PIPELINE_URL",

		"CI_JOB_ID", "CI_JOB_NAME", "CI_JOB_STAGE", "CI_JOB_URL",

		// "CI_BUILD_ID", "CI_BUILD_REF", "CI_BUILD_NAME", "CI_BUILD_REF_SLUG",
		// "CI_BUILD_REF_NAME", "CI_BUILD_STAGE"
	}
	keys = append(keys, ciKeys...)

	return getOsEnvs(keys)
}

// toContextMap converts a Gitlab env map to ContextMap object
// This is due to map type casting
/*
func (env GitlabEnv) toContextMap() *ContextMap {
	cm := make(ContextMap, len(env))
	for k, v := range env {
		cm[k] = v
	}
	return &cm
}
*/

// getOsEnvs returns a map of env variables, from a list of targeted keys
// only if the key is defined as env var; otherwise key is ignored
func getOsEnvs(keys []string) map[string]string {
	env := make(map[string]string)
	for _, key := range keys {
		// env[key] = os.Getenv(key)
		if val, found := os.LookupEnv(key); found {
			env[key] = val
		}
	}
	return env
}

// IsCI indicates whether the CLI is run from a CI platform
func IsCI() bool {
	return os.Getenv("CI") == "true"
}

// IsGithubCI indicates whether the CLI is run from a Github worflow
func IsGithubCI() bool {
	return IsCI() && os.Getenv("GITHUB_ACTIONS") == "true"
}

// IsGitlabCI indicates whether the CLI is run from a Gitlab pipeline job
func IsGitlabCI() bool {
	return IsCI() && os.Getenv("GITLAB_CI") == "true"
}

// IsGithubRepo indicates whether the CLI is run against a github repo
func IsGithubRepo() bool {
	_, ok := os.LookupEnv("GITHUB_REPOSITORY")
	return ok
}

// IsGitlabRepo indicates whether the CLI is run against a gitlab repo
func IsGitlabRepo() bool {
	_, ok := os.LookupEnv("CI_PROJECT_NAME")
	return ok
}

// NewContext is a constructor to initialize a full context for Reliably
func NewContext() *Context {
	var (
		ci bool = IsCI()

		ctxMap  ContextMap = map[string]interface{}{}
		environ StringMap  = map[string]string{}
	)

	if IsGithubCI() {
		environ = StringMap(GetGithubEnv())
	} else if IsGitlabCI() {
		environ = StringMap(GetGitlabEnv())
	}

	context := &Context{
		IsCI:    ci,
		Runtime: NewRuntimeContext(),
		Context: ctxMap,
		Environ: environ,
		Source:  nil,
	}

	// dynamically sets the source , based on runtime & environ
	setSource(context)

	return context
}

type SourceType uint

const ( // iota is reset to 0 - is like an enum
	local SourceType = iota
	github
	gitlab
	git
)

var SourceTypeToString = map[SourceType]string{
	local:  "local",
	github: "GitHub",
	gitlab: "GitLab",
	git:    "git", // git repository
}

func (st SourceType) String() string {
	if s, ok := SourceTypeToString[st]; ok {
		return s
	}
	return "unknown"
}

func (st SourceType) MarshalJSON() ([]byte, error) {
	if s, ok := SourceTypeToString[st]; ok {
		return json.Marshal(s)
	}
	return nil, fmt.Errorf("unknown user type %d", st)
}

type SourceMeta map[string]string

// Source is the base interface for a source files to be checked by CLI
type Source struct {
	Type SourceType `json:"type"`
	Hash string     `json:"hash"`
	Meta SourceMeta `json:"meta"`
}

// LocalSource represents a local folder
type LocalSource struct {
	Type SourceType `json:"type"`
}

// GithubSource represents a GitHub repository
type GithubSource struct {
	Type SourceType `json:"type"`
}

// GitlabSource represents a Gitlab project
type GitlabSource struct {
	Type SourceType `json:"type"`
}

// setSource updates the context with the source, that is scanned by the CLI
// It inditates the type of the source code and meta to be able to uniquely
// indentify it, as well as to refer back to it
func setSource(context *Context) {
	var source *Source
	var err error

	if utils.IsGitRepo() {

		url, _ := utils.GitRemoteOriginURL()
		source, err = NewGitSource(url)
		if err != nil {
			return
		}

	} else if IsGithubRepo() {
		// in case the git repo has not been checked out in the GitHub workflow job
		serverURL := context.Environ["GITHUB_SERVER_URL"]
		repo := context.Environ["GITHUB_REPOSITORY"]
		url := fmt.Sprintf("%s/%s", serverURL, repo)
		source, err = NewGitSource(url)
		if err != nil {
			return
		}

	} else if IsGitlabRepo() {
		// in case the git repo has not been checked out in the GitLab pipeline job
		projectURL := context.Environ["CI_PROJECT_URL"]
		source, err = NewGitSource(projectURL)
		if err != nil {
			return
		}

	} else {
		source = &Source{
			Type: local,
			Meta: map[string]string{
				"hostname":   context.Runtime.Hostname,
				"workingdir": context.Runtime.WorkingDir,
			},
		}
	}

	context.Source = *source
}

func NewGitSource(url string) (source *Source, err error) {

	gURL, err := utils.ParseGitRemoteOriginURL(url)
	if err != nil {
		return
	}

	source = &Source{
		Type: git,
		Hash: gURL.Hash(),
		Meta: map[string]string{
			"server": gURL.Host,
			"repo":   gURL.Path,
		},
	}

	return
}
