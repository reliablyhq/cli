package report

import (
	"github.com/sirupsen/logrus"
)

const (
	threshold                       = 0
	lessThan95pcAvailabilityMessage = "An availability of less than 95% allows more than 36.5 hours of downtime per month, which should be possible for any well built app deployed as a single instance. This availability target is probably not high enough for a production-ready system."
	errorBudgetExceededf            = "Tour error budget has been exceeded by %.2f percent. This is pretty bad :("
	errorBudgetTooLowf              = "You are under your error budget by %.2f percent. You could tighten your budget, or could decrease the quality of the experience your application provides (e.g by reducing the amount of resources given to your application)."
	latencyExceeded                 = "The average latency threshold has been exceeeded by %vms"
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

	if r.ServiceLevel.Delta.ErrorPercent < threshold {
		l.Warnf(errorBudgetTooLowf, -r.ServiceLevel.Delta.ErrorPercent)
	} else if r.ServiceLevel.Delta.ErrorPercent > threshold {
		l.Warnf(errorBudgetExceededf, r.ServiceLevel.Delta.ErrorPercent)
	}

	if r.ServiceLevel.Delta.LatencyMs > threshold {
		l.Warnf(latencyExceeded, r.ServiceLevel.Delta.LatencyMs)
	}
}
