package report

import (
	"github.com/sirupsen/logrus"
)

const (
	lessThan95pcAvailabilityMessage = "An availability of less than 95% allows more than 36.5 hours of downtime per month, which should be possible for any well built app deployed as a single instance. This availability target is probably not high enough for a production-ready system."
	actualAvailabilityTooLowf       = "Current availability is lower than target availability by %.2f percent. Think about increasing the resources allocated of your application"
	actualAvailabilityTooHighf      = "Current availability is higher than target availability by %.2f percent. Think about reducing the resources allocated to your application - this could save you some money."
	errorBudgetExceededf            = "Error budget has been exceeded by %.2f percent. This is pretty bad :("
	errorBudgetTooLowf              = "You are under your error budget by %.2f percent. You could tighten your budget, or could decrease the quality of the experience your application provides (e.g by reducing the amount of resources given to your application)."
	latencyExceeded                 = "The latency threshold has been exceeeded by %vms"
)

func Write(r *Report, l *logrus.Logger) {
	if r == nil {
		return
	}

	if l == nil {
		return
	}

	if r.Delta == nil {
		l.Error("the report does not include a 'Delta'")
		return
	}

	if r.Delta.ServiceLevel < 0 { // low service level
		l.Warnf(actualAvailabilityTooLowf, -r.Delta.ServiceLevel)
	} else if r.Delta.ServiceLevel > 2 {
		l.Warnf(actualAvailabilityTooHighf, r.Delta.ServiceLevel)
	}

	if r.Delta.ErrorBudgetPercent < -2 {
		l.Warnf(errorBudgetTooLowf, r.Delta.ErrorBudgetPercent)
	} else if r.Delta.ErrorBudgetPercent > 0 {
		l.Warnf(errorBudgetExceededf, r.Delta.ErrorBudgetPercent)
	}

	if r.Delta.LatencyMs > 0 {
		l.Warnf(latencyExceeded, r.Delta.LatencyMs)
	}
}
