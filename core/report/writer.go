package report

import (
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
)

const (
	lessThan95pcAvailabilityMessage = "An availability of less than 95% allows more than 36.5 hours of downtime per month, which should be possible for any well built app deployed as a single instance. This availability target is probably not high enough for a production-ready system."
	actualAvailabilityTooLowf       = "Current availability is lower than target availability by %.2f percent. Think about increasing the resources allocated of your application"
	actualAvailabilityTooHighf      = "Current availability is higher than target availability by %.2f percent. Think about reducing the resources allocated to your application - this could save you some money."
	errorBudgetExceededf            = "Error budget has been exceeded by %.2f percent. This is pretty bad :("
	errorBudgetTooLowf              = "You are under your error budget by %.2f percent. You could tighten your budget, or could decrease the quality of the experience your application provides (e.g by reducing the amount of resources given to your application)."
)

func Write(r *Report, l *logrus.Logger) {
	if l == nil {
		return
	}

	if r.Delta.ServiceLevelPercent > r.Threshold.Error {
		log.Errorf(actualAvailabilityTooHighf, r.Delta.ServiceLevelPercent)
	} else if r.Delta.ServiceLevelPercent > r.Threshold.Warning {
		log.Warnf(actualAvailabilityTooHighf, r.Delta.ServiceLevelPercent)
	} else if r.Delta.ServiceLevelPercent < 95 {
		log.Warn(lessThan95pcAvailabilityMessage)
	}

	// todo: write more stuff
}
