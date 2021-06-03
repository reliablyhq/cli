package initAlpha

import (
	"errors"

	initCmd "github.com/reliablyhq/cli/cmd/reliably/slo/init"
)

func AlpaInitRun(opts *initCmd.InitOptions) error {
	return errors.New("This is alpha version of the 'slo init' command")
}
