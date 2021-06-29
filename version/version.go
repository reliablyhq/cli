package version

import (
	"runtime/debug"

	"github.com/reliablyhq/cli/utils"
)

const DevVersion = "DEV"

// Version of the CLI with semver format
// Version is dynamically set by the toolchain or overridden by the Makefile.
var Version = DevVersion

// Date is dynamically set at build time in the Makefile.
var Date = "" // YYYY-MM-DD

// Revision is dynamically set at build time in the Makefile
var Revision = ""

func init() {

	if Version == "" {
		Version = DevVersion
	}

	if Version == DevVersion {
		if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "(devel)" && info.Main.Version != "" {
			Version = info.Main.Version
		}

		if Revision == "" {
			// sets by default at runtime the revision, if not configured by ldflag
			gitRev, err := utils.GitShortRev()
			if err == nil {
				Revision = gitRev
			}
		}
	}
}

// IsDevVersion indicates whether the current version is DEV
func IsDevVersion() bool {
	return Version == DevVersion
}
