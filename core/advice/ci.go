package advice

import (
	"strings"
)

func getAdviceForCI(ci string) ([]*Advice, error) {
	switch strings.ToLower(ci) {
	default:
		{
			return nil, notSuportedErrorBuilder("CI", ci)
		}
	}
}
