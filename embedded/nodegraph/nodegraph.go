package nodegraph

import "embed"

//go:embed static
var FS embed.FS

// RootDir - the root directory of the FileSystem
const RootDir = "static"
