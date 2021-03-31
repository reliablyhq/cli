package init

import (
	"os"
	"os/user"
	"path/filepath"

	"github.com/reliablyhq/cli/utils"
)

func getDefaultAppName() string {

	if utils.IsGitRepo() {
		if url, err := utils.GitRemoteOriginURL(); err == nil {
			_, repo, err := utils.ExtractOwnerRepoFromGitURL(url)
			if err == nil {
				return repo
			}
		}
	}

	if cwd, err := os.Getwd(); err == nil {
		return filepath.Base(cwd)
	}

	return "my-app"
}

func getDefaultAppOwner() string {

	if utils.IsGitRepo() {
		if url, err := utils.GitRemoteOriginURL(); err == nil {
			owner, _, err := utils.ExtractOwnerRepoFromGitURL(url)
			if err == nil {
				return owner
			}
		}
	}

	user, _ := user.Current()
	return user.Username
}

func getDefaultRepository() string {

	if utils.IsGitRepo() {
		if url, err := utils.GitRemoteOriginURL(); err == nil {
			return url
		}
	}

	cwd, _ := os.Getwd()
	return cwd
}
