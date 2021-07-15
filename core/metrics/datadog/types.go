package datadog

import (
	"fmt"
	"time"

	"github.com/reliablyhq/cli/core/entities"
)

type Datadog struct {
}

// NewGCP currently needs a service account configured or gcloud auth within same session to function
func NewDatadog() (*Datadog, error) {
	return &Datadog{}, nil
}

// CanHandleSelector returns true only if both datadog queries are defined
func (dd *Datadog) CanHandleSelector(labels entities.Selector) bool {

	if val, ok := labels["datadog_numerator_query"]; !ok || val == "" {
		return false
	}
	if val, ok := labels["datadog_denominator_query"]; !ok || val == "" {
		return false
	}

	return true
}

func (dd *Datadog) ComputeObjective(o *entities.Objective, from time.Time, to time.Time) (*entities.Indicator, error) {

	if ok, err := ValidateApiKey(); !ok {
		return nil, fmt.Errorf("Error while validating DD API KEY: %s", err)
	}

	i := entities.NewIndicatorForObjective(o, from, to)

	num_query := o.Spec.IndicatorSelector["datadog_numerator_query"]
	denom_query := o.Spec.IndicatorSelector["datadog_denominator_query"]

	slo, err := ComputeSloFromQueryMetrics(num_query, denom_query, from, to)
	if err == nil {
		i.Spec.Percent = slo
	}

	return i, nil
}

// No resource identifier can be found, this is part of DD env vars
// as we are running queries to datadog api
func (dd *Datadog) ResourceFromSelector(labels entities.Selector) string {
	return ""
}

// Nothing to do for Datadog...
func (dd *Datadog) Close() error {
	return nil
}
