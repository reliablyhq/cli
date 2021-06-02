package workflows

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/MakeNowJust/heredoc/v2"
	"gopkg.in/yaml.v2"

	"github.com/reliablyhq/cli/utils"
)

type workflow map[string]interface{}

// GetWorkflowPath returns the path of the workflow file to generate
// for the specified CI/CD platform
func GetWorkflowPath(platform string) string {

	var p string

	switch platform {
	case "github":
		p = github_Path
	case "gitlab":
		p = gitlab_Path
	case "circleci":
		p = circleci_Path
	default:
		p = ""
	}

	return p
}

// GetWorflow returns the default workflow content for the CI/CD platform
func GetWorkflow(platform string) string {
	var tpl string

	switch platform {
	case "github":
		tpl = github_Template
	case "gitlab":
		tpl = gitlab_Template
	case "circleci":
		tpl = circleci_Template
	default:
		tpl = ""
	}

	return tpl
}

// Generates a default worfklow file for the specified CI/CD platform,
// at a default path (recommanded by the platform)
// If a workflow file already exists at default file location,
// we try to update it or return an error
func GenerateWorkflow(platform string) (string, error) {

	p := GetWorkflowPath(platform)
	c := GetWorkflow(platform)

	if _, err := os.Stat(p); !os.IsNotExist(err) {
		err := editWorkflowInPlace(platform)
		return p, err
	}

	err := mkWorkflowDirAll(p)
	if err != nil {
		return p, nil
	}

	err = writeWorkflowToFile(p, c)
	if err != nil {
		return p, err
	}

	return p, nil
}

// workflowDir returns the directory path that will contain the workflow file
func workflowDir(path string) string {
	if !strings.Contains(path, "/") {
		return ""
	}

	return filepath.Dir(path)
}

// mkWorkflowDirAll ensure to create nested directories for the full
// workflow file path, if they do not exist
func mkWorkflowDirAll(path string) error {

	dir := workflowDir(path)
	if dir != "" {
		err := os.MkdirAll(dir, 0766)
		if err != nil {
			return err
		}
	}

	return nil
}

func writeWorkflowToFile(path string, content string) error {

	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("Unable to open the file at path: %s", path)
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	_, err = writer.WriteString(content)
	if err != nil {
		return fmt.Errorf("Unable to write to file at path: %s", path)
	}
	writer.Flush()

	return nil
}

// loadWorkflow loads the workflow from raw string into a yaml compliant map
func loadWorkflow(raw string) (workflow, error) {
	wf := make(workflow)

	err := yaml.Unmarshal([]byte(raw), &wf)
	if err != nil {
		return nil, err
	}

	return wf, nil
}

// loadWorkflowFile returns the loaded workflow into yaml compliant map
// from a local file at a given path
// useful to laod into memory the workflow, before editing/inserting
// the reliably workflow parts
func loadWorkflowFile(path string) (workflow, error) {

	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	return loadWorkflow(string(raw))
}

// canEditWorkflowInPlace indicates whether the workflow can be inserted
// within an existing CI/CD workflow file
func CanEditWorkflowInPlace(platform string) bool {
	switch platform {
	case "gitlab", "circleci":
		return true
	}
	return false
}

// EditWorkflowInPlace inserts the Reliably workflow parts into an existing
// CI/CD workflow file
func editWorkflowInPlace(platform string) error {
	switch platform {
	case "gitlab":
		return editWorkflow(gitlab_Template, gitlab_Path, insertReliablyToGitlab)
	case "circleci":
		return editWorkflow(circleci_Template, circleci_Path, insertReliablyToCircleci)
	}
	return fmt.Errorf("Unable to edit the workflow file in place for %s", platform)
}

// marshalWorkflow returns a yaml formatted string of a workflow map
func marshalWorkflow(wf workflow) (string, error) {
	s, err := yaml.Marshal(&wf)
	if err != nil {
		return "", err
	}
	return string(s), nil
}

type editWorkflowFunc func(wf workflow, tpl workflow) error

// editWorkflow allows to inject reliably workflow parts into an existing
// CI/CD workflow file
// It takes the default template, the existing workflow path and an
// updater function that has a specific signature
// The updater function is responsible to modify the workflow map
// If the update does not seem possible, it returns an error that will be
// returned to the user
func editWorkflow(template string, path string, f editWorkflowFunc) error {

	tpl, err := loadWorkflow(template)
	if err != nil {
		return err
	}

	wf, err := loadWorkflowFile(path)
	if err != nil {
		return err
	}

	err = f(wf, tpl) // call the workflow edition function
	if err != nil {
		return err
	}

	content, err := marshalWorkflow(wf)
	if err != nil {
		return err
	}

	err = writeWorkflowToFile(path, content)
	if err != nil {
		return err
	}

	return nil
}

// GetAccessTokenHelp returns the help message on how to securely
// setup the reliably access token required for CLI to make authenticated
// calls to the API
func GetAccessTokenHelp(platform string, envvarname string) (help string) {

	switch platform {
	case "github":
		owner, repo := "OWNER", "REPO"

		url, err := utils.GitRemoteOriginURL()
		if err == nil && strings.Contains(url, "github.com") {
			if o, r, err := utils.ExtractOwnerRepoFromGitURL(url); err == nil {
				owner, repo = o, r
			}
		}

		help = heredoc.Docf(github_AccessTokenSecretHelp, envvarname, owner, repo)
	case "gitlab":
		owner, repo := "OWNER", "PROJECT"

		url, err := utils.GitRemoteOriginURL()
		if err == nil && strings.Contains(url, "gitlab.com") {
			if o, r, err := utils.ExtractOwnerRepoFromGitURL(url); err == nil {
				owner, repo = o, r
			}
		}

		help = heredoc.Docf(gitlab_AccessTokenHelp, envvarname, owner, repo)

	case "circleci":
		help = heredoc.Docf(circleci_AccessTokenHelp, envvarname)
	default:
		help = ""
	}

	return
}
