package gcp

import (
	"context"
	"fmt"
	"strings"
	"time"

	monitoring "cloud.google.com/go/monitoring/apiv3"
	"github.com/golang/protobuf/ptypes/duration"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/reliablyhq/cli/core/metrics"
	"google.golang.org/api/iterator"
	monitoringpb "google.golang.org/genproto/googleapis/monitoring/v3"
)

type GCP struct {
	ctx    context.Context
	client *monitoring.MetricClient
}

// type checking
var _ metrics.Provider = &GCP{}

// Currently needs a service account configured or gcloud auth within same session
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

func (p *GCP) Get99PercentLatencyMetricForResource(resourceID string, from, to time.Time) (float64, error) {

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

	latency, err := parseLatency(timeSeries[0])
	if err != nil {
		return -1, err
	}

	return latency, nil
}

func (p *GCP) GetErrorPercentageMetricForResource(resourceID string, from, to time.Time) (float64, error) {

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
		return -1, err
	}

	return errorPercentage, nil
}

func parseLatency(resp *monitoringpb.TimeSeries) (float64, error) {

	latency := resp.GetPoints()[0].GetValue().GetDoubleValue()

	return latency, nil
}

func parseStatusErrors(it []*monitoringpb.TimeSeries) (float64, error) {

	responseTotal := 0
	errorCount := 0
	var errorPercentage float64 = 0

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
