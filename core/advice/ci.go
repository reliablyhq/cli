package advice

import (
	"errors"

	"github.com/reliablyhq/cli/manifest"
)

func getSuggestionsForCI(ci *manifest.ContinuousIntegrationInfo) ([]Suggestion, error) {
	return nil, errors.New("getSuggestionForCI has not been implemented")
}
