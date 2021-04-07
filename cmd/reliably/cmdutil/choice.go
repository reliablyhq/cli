package cmdutil

import (
	"github.com/reliablyhq/cli/utils"
)

// Choice is a list of string
type Choice []string

// Has indicates whether the string slice contains the value
func (list Choice) Has(a string) bool {
	return utils.StringInArray(a, list)
}
