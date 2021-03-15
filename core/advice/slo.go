package advice

import (
	"errors"
	"time"
)

type Downtime struct {
	Day     time.Duration
	Week    time.Duration
	Month   time.Duration
	Quarter time.Duration
	Year    time.Duration
}

func getAdviceForSLO(availability float32) ([]*Advice, error) {
	return nil, errors.New("getAdviceForSLO has not been implemented")
}

func calculateDowntime(availability float32) *Downtime {
	dt := Downtime{}

	return &dt
}
