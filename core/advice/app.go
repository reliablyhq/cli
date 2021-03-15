package advice

import (
	"strings"

	"github.com/reliablyhq/cli/core"
)

func getAdviceForType(t core.AppType) ([]*Advice, error) {
	tString := string(t)
	switch strings.ToLower(tString) {
	default:
		{
			return nil, notSuportedErrorBuilder("Application", tString)
		}
	}
}
