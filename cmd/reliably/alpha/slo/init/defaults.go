package initAlpha

import (
	"fmt"
	//"os"
	//"os/user"
	//"path/filepath"
	"strings"

	"github.com/reliablyhq/cli/core/entities"
	//"github.com/reliablyhq/cli/utils"
	"github.com/reliablyhq/cli/core"
	//"github.com/reliablyhq/cli/core/manifest"
)

/*
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
*/
/*
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
*/

/*
func getDefaultRepository() string {

	if utils.IsGitRepo() {
		if url, err := utils.GitRemoteOriginURL(); err == nil {
			return url
		}
	}

	cwd, _ := os.Getwd()
	return cwd
}
*/

/*
func initDefaultSloName(sl *manifest.ServiceLevel) error {

	var desc string
	switch sl.Type {
	case "latency":
		c := sl.Criteria.(manifest.LatencyCriteria)
		threshold := c.Threshold.Duration.Milliseconds()
		desc = fmt.Sprintf("faster than %vms", threshold)
	case "availability":
		desc = "successful"
	}

	sl.Name = fmt.Sprintf("%v%% of requests %s over last %s",
		sl.Objective,
		desc,
		core.HumanizeDuration(sl.ObservationWindow.ToDuration()),
	)

	return nil
}
*/

func generateDefaultSloName(o entities.Objective) string {
	var (
		name     string
		desc     string
		category string
	)

	if val, ok := o.Spec.IndicatorSelector["category"]; ok {
		category = val
	}

	switch category {
	case "latency":
		threshold := o.Spec.IndicatorSelector["latency_target"]

		if !strings.HasSuffix(threshold, "ms") {
			threshold = fmt.Sprintf("%sms", threshold)
		}

		desc = fmt.Sprintf("faster than %s", threshold)

	case "availability":
		desc = "successful"
	}

	name = fmt.Sprintf("%v%% of requests %s over last %s",
		o.Spec.ObjectivePercent,
		desc,
		core.HumanizeDuration(o.Spec.Window.Duration),
	)

	return name
}
