package advice

import (
	"fmt"
	"strings"

	"github.com/reliablyhq/cli/core"
)

func GetAdviceFor(m *core.Manifest) ([]*Advice, error) {
	var err error
	allAdvice := make([]*Advice, 0)

	switch strings.ToLower(m.Type) {
	default:
		{
			err = fmt.Errorf("Application type '%s' is not supported", m.Type)
		}
	}

	return allAdvice, err
}
