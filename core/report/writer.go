package report

import (
	"github.com/sirupsen/logrus"
)

const (
	threshold                       = 0
	lessThan95pcAvailabilityMessage = "An availability of less than 95% allows more than 36.5 hours of downtime per month, which should be possible for any well built app deployed as a single instance. This availability target is probably not high enough for a production-ready system."
	actualAvailabilityTooLowf       = "Current availability is lower than target availability by %.2f percent. Think about increasing the resources allocated of your application"
	actualAvailabilityTooHighf      = "Current availability is higher than target availability by %.2f percent. Think about reducing the resources allocated to your application - this could reduce operating costs."
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

	if r.ServiceLevel.Delta == nil {
		l.Error("the report does not include a 'Delta'")
		return
	}

	if r.ServiceLevel.Delta.ErrorBudgetPercent < threshold {
		l.Warnf(errorBudgetTooLowf, -r.ServiceLevel.Delta.ErrorBudgetPercent)
	} else if r.ServiceLevel.Delta.ErrorBudgetPercent > threshold {
		l.Warnf(errorBudgetExceededf, r.ServiceLevel.Delta.ErrorBudgetPercent)
	}

	if r.ServiceLevel.Delta.LatencyMs > threshold {
		l.Warnf(latencyExceeded, r.ServiceLevel.Delta.LatencyMs)
	}
}
