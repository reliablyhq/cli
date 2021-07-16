package gcp

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	monitoring "cloud.google.com/go/monitoring/apiv3/v2"
	"github.com/golang/protobuf/ptypes/duration"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/reliablyhq/cli/core/entities"
	log "github.com/sirupsen/logrus"
	"google.golang.org/api/iterator"
	monitoringpb "google.golang.org/genproto/googleapis/monitoring/v3"
)

type GCP struct {
	ctx    context.Context
	client *monitoring.MetricClient
}

// NewGCP currently needs a service account configured or gcloud auth within same session to function
func NewGCP() (*GCP, error) {

	ctx := context.Background()
	client, err := monitoring.NewMetricClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("NewMetricClient: %v", err)
	}

	return &GCP{
		ctx:    ctx,
		client: client,
	}, nil
}

// Close closes the GCP monitoring client connection
// See: https://pkg.go.dev/cloud.google.com/go@v0.81.0/monitoring/apiv3/v2#MetricClient.Close
func (p *GCP) Close() error {
	return p.client.Close()
}

// Get99PercentLatencyMetricForResource retrieves latency data for a resource on GCP
// and returns the mean of the 99th Percentile latencies across regions of the given resource
func (p *GCP) Get99PercentLatencyMetricForResource(resourceID string, from, to time.Time) (float64, error) {

	if !isValidResourceID(resourceID) {
		return -1, fmt.Errorf("Resource ID not valid: %s", resourceID)
	}
	log.Debugf("Resource Id: %#v", resourceID)
	log.Debugf("Retrieve latency metrics From %s To %s", from, to)

	projectID := strings.SplitN(resourceID, "/", -1)[0]

	metricType := "loadbalancing.googleapis.com/https/backend_latencies"
	resourceName := strings.SplitN(resourceID, "/", -1)[2]
	alignmentSeconds := to.Unix() - from.Unix()

	req := &monitoringpb.ListTimeSeriesRequest{
		Name: "projects/" + projectID,
		Filter: fmt.Sprintf(
			`metric.type="%s" AND resource.labels.url_map_name="%s"`,
			metricType,
			resourceName,
		),
		Interval: &monitoringpb.TimeInterval{
			StartTime: &timestamp.Timestamp{
				Seconds: from.Unix(),
			},
			EndTime: &timestamp.Timestamp{
				Seconds: to.Unix(),
			},
		},
		Aggregation: &monitoringpb.Aggregation{
			CrossSeriesReducer: monitoringpb.Aggregation_REDUCE_MEAN,
			PerSeriesAligner:   monitoringpb.Aggregation_ALIGN_PERCENTILE_99,
			AlignmentPeriod: &duration.Duration{
				Seconds: alignmentSeconds,
			},
		},
	}
	timeSeries, err := collectIterations(p.client.ListTimeSeries(p.ctx, req))
	if err != nil {
		return -1, err
	}

	latency, err := parseLatency(timeSeries)
	if err != nil {
		return -1, fmt.Errorf("%s: %s", err, resourceID)
	}

	log.Debugf("99 percentile latency is %.3fms\n", latency)

	return latency, nil
}

func (p *GCP) GetLatencyAboveThresholdPercentage(resourceID string, from, to time.Time, threshold int) (float64, error) {

	if !isValidResourceID(resourceID) {
		return -1, fmt.Errorf("Resource ID not valid: %s", resourceID)
	}
	log.Debugf("Resource Id: %#v", resourceID)
	log.Debugf("Retrieve latency metrics From %s To %s", from, to)

	projectID := strings.SplitN(resourceID, "/", -1)[0]

	metricType := "loadbalancing.googleapis.com/https/total_latencies"
	resourceName := strings.SplitN(resourceID, "/", -1)[2]

	req := &monitoringpb.ListTimeSeriesRequest{
		Name: "projects/" + projectID,
		Filter: fmt.Sprintf(
			`metric.type="%s" AND resource.labels.url_map_name="%s"`,
			metricType,
			resourceName,
		),
		Interval: &monitoringpb.TimeInterval{
			StartTime: &timestamp.Timestamp{
				Seconds: from.Unix(),
			},
			EndTime: &timestamp.Timestamp{
				Seconds: to.Unix(),
			},
		},
		Aggregation: &monitoringpb.Aggregation{
			PerSeriesAligner: monitoringpb.Aggregation_ALIGN_PERCENTILE_99,
			AlignmentPeriod: &duration.Duration{
				Seconds: 60,
			},
			GroupByFields: []string{"resource.label.url_map_name"},
		},
	}
	timeSeries, err := collectIterations(p.client.ListTimeSeries(p.ctx, req))
	if err != nil {
		return -1, err
	}

	latencyThresholdPercentage, err := calculateLatencyOverThresholdPercentage(threshold, timeSeries)
	if err != nil {
		return -1, err
	}

	log.Debugf("Latency SLI is %.3f%%\n", latencyThresholdPercentage)

	return latencyThresholdPercentage, nil
}

// GetErrorPercentageMetricForResource retrieves the error status code data for a resource on
// GCP and calculates percentage of 500 status code
func (p *GCP) GetErrorPercentageMetricForResource(resourceID string, from, to time.Time) (float64, error) {

	if !isValidResourceID(resourceID) {
		return -1, fmt.Errorf("Resource ID not valid: %s", resourceID)
	}
	log.Debugf("Resource Id: %#v", resourceID)
	log.Debugf("Retrieve error rate metrics From %s To %s", from, to)

	projectID := strings.SplitN(resourceID, "/", -1)[0]

	metricType := "loadbalancing.googleapis.com/https/request_count"
	resourceName := strings.SplitN(resourceID, "/", -1)[2]
	alignmentSeconds := to.Unix() - from.Unix()

	req := &monitoringpb.ListTimeSeriesRequest{
		Name: "projects/" + projectID,
		Filter: fmt.Sprintf(
			`metric.type="%s" AND resource.labels.url_map_name="%s" AND metric.label.response_code_class != 0`,
			metricType,
			resourceName,
		),
		Interval: &monitoringpb.TimeInterval{
			StartTime: &timestamp.Timestamp{
				Seconds: from.Unix(),
			},
			EndTime: &timestamp.Timestamp{
				Seconds: to.Unix(),
			},
		},
		Aggregation: &monitoringpb.Aggregation{
			CrossSeriesReducer: monitoringpb.Aggregation_REDUCE_SUM,
			PerSeriesAligner:   monitoringpb.Aggregation_ALIGN_SUM,
			AlignmentPeriod: &duration.Duration{
				Seconds: alignmentSeconds,
			},
			GroupByFields: []string{"metric.label.response_code_class"},
		},
	}

	timeSeries, err := collectIterations(p.client.ListTimeSeries(p.ctx, req))
	if err != nil {
		return -1, err
	}

	errorPercentage, err := parseStatusErrors(timeSeries)
	if err != nil {
		return -1, fmt.Errorf("%s: %s", err, resourceID)
	}

	log.Debugf("error rate is %.2f%%\n", errorPercentage)
	return errorPercentage, nil
}

func (p *GCP) GetAvailabilityPercentage(resourceID string, from, to time.Time) (float64, error) {
	errorRate, err := p.GetErrorPercentageMetricForResource(resourceID, from, to)
	if err != nil {
		return -1, err
	}

	availability := 100.0 - errorRate
	log.Debugf("Availability is %.2f%%\n", availability)
	return availability, nil
}

// ResourceFromSelector - identifies the resource ID given a selector.
func (p *GCP) ResourceFromSelector(s entities.Selector) string {
	if pid, ok := s["gcp_project_id"]; ok {
		// check for loadbalancer key
		if lb, ok := s["gcp_loadbalancer_name"]; ok {
			return fmt.Sprintf("%s/google-cloud-load-balancers/%s", pid, lb)
		}
	}

	return ""
}

func calculateLatencyOverThresholdPercentage(threshold int, it []*monitoringpb.TimeSeries) (float64, error) {
	if len(it) < 1 {
		return -1, fmt.Errorf("No data found for resource")
	}

	responsesUnderThreshold := 0
	totalResponses := 0
	for _, resp := range it {
		latency := resp.GetPoints()[0].GetValue().GetDoubleValue()
		totalResponses += 1
		if latency < float64(threshold) {
			responsesUnderThreshold += 1
		}
	}

	latencyPercentage := (float64(responsesUnderThreshold) / float64(totalResponses)) * 100

	return latencyPercentage, nil
}

func parseLatency(it []*monitoringpb.TimeSeries) (float64, error) {

	if len(it) < 1 {
		return -1, fmt.Errorf("No data found for resource")
	}

	latency := it[0].GetPoints()[0].GetValue().GetDoubleValue()

	return latency, nil
}

func parseStatusErrors(it []*monitoringpb.TimeSeries) (float64, error) {

	if len(it) < 1 {
		return -1, fmt.Errorf("No data found for resource")
	}

	responseTotal := 0
	errorCount := 0
	var errorPercentage float64

	for _, resp := range it {
		if resp.GetMetric().GetLabels()["response_code_class"] == "500" {
			errorCount = int(resp.GetPoints()[0].GetValue().GetInt64Value())
		}

		responseTotal += int(resp.GetPoints()[0].GetValue().GetInt64Value())
	}
	if errorCount != 0 {
		errorPercentage = (float64(errorCount) / float64(responseTotal)) * 100
	} else {
		errorPercentage = 0
	}

	return errorPercentage, nil
}

func collectIterations(it *monitoring.TimeSeriesIterator) ([]*monitoringpb.TimeSeries, error) {
	var result []*monitoringpb.TimeSeries

	for {
		resp, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("Could not read time series value: %v", err)
		}

		result = append(result, resp)
	}

	return result, nil
}

func isValidResourceID(resourceID string) bool {
	if len(strings.SplitN(resourceID, "/", -1)) < 3 {
		return false
	}
	return true

}

func (p *GCP) CanHandleSelector(labels entities.Selector) bool {
	resourceID := p.ResourceFromSelector(labels)
	if resourceID == "" {
		return false
	}

	return true
}

func (p *GCP) ComputeObjective(o *entities.Objective, from time.Time, to time.Time) (*entities.Indicator, error) {

	i := entities.NewIndicatorForObjective(o, from, to)
	var err error

	resourceID := p.ResourceFromSelector(o.Spec.IndicatorSelector)
	if resourceID == "" {
		return nil, fmt.Errorf("unable to identify provider and resource id for objective: %v",
			o.Spec.IndicatorSelector)
	}

	switch o.Spec.IndicatorSelector["category"] {
	case "latency":
		target, ok := o.Spec.IndicatorSelector["latency_target"]
		if !ok {
			return nil, errors.New("latency_target not defined in Objective spec")
		}

		thres, err := time.ParseDuration(target)
		if err != nil {
			return nil, err
		}

		i.Spec.Percent, err = p.GetLatencyAboveThresholdPercentage(
			resourceID, from, to, int(thres.Milliseconds()))
		if err != nil {
			return nil, err
		}

	case "availability":
		i.Spec.Percent, err = p.GetAvailabilityPercentage(
			resourceID, from, to)
		if err != nil {
			return nil, err
		}

	default:
		return nil, fmt.Errorf("unsupported indicator category: %s",
			o.Spec.IndicatorSelector["category"])
	}

	return i, nil
}
