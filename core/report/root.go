package report

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/reliablyhq/cli/errors"
	"github.com/reliablyhq/cli/manifest"
)

const (
	oneDay  = 24 * time.Hour
	oneWeek = 7 * oneDay
)

var (
	r = rand.New(rand.NewSource(time.Now().Unix()))
)

var notSuportedErrorBuilder = func(thingType, thingName string) error {
	return fmt.Errorf("%s type '%s' is not currently supported. I've informed ReliablyHQ about this; check back later - maybe we'll be able to help then.", thingType, thingName)
}

func GetAdviceFor(m *manifest.Manifest) (*Report, error) {
	r := Report{}
	allErrors := make([]error, 0)

	if d, err := getCurrentErrorPc(m, oneWeek); err == nil {
		r.Delta.ErrorBudgetPercent = m.ServiceLevel.ErrorBudgetPercent - d
	} else {
		allErrors = append(allErrors, err)
	}

	if d, err := getCurrentAvailability(m, oneWeek); err == nil {
		r.Delta.ServiceLevelPercent = m.ServiceLevel.Availability - d
	} else {
		allErrors = append(allErrors, err)
	}

	if d, err := get99PercentLatency(m, oneWeek); err == nil {
		r.Delta.LatencyCeilingPercent = float32(d / m.ServiceLevel.Latency * 100)
	} else {
		allErrors = append(allErrors, err)
	}

	if len(allErrors) > 0 {
		return &r, errors.NewCompoundError("multiple errors occured", allErrors)
	}

	return &r, nil
}
