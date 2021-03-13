package advice

import (
	"errors"

	"github.com/reliablyhq/cli/core"
)

func GetAdviceForPlatform(p string, m *core.Manifest) (*Advice, error) {
	// todo: get this remotely

	return nil, errors.New("not implemented")
}
