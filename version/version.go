package version

import (
	"runtime/debug"
)

const devVersion = "DEV"

// Version of the CLI with semver format
// Version is dynamically set by the toolchain or overridden by the Makefile.
var Version = devVersion

// Date is dynamically set at build time in the Makefile.
var Date = "" // YYYY-MM-DD

func init() {

	if Version == "" {
		Version = devVersion
	}

	if Version == devVersion {
		if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "(devel)" && info.Main.Version != "" {
			Version = info.Main.Version
		}
	}
}