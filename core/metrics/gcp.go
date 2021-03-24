package metrics

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	monitoring "cloud.google.com/go/monitoring/apiv3"
	"github.com/golang/protobuf/ptypes/duration"
	"github.com/golang/protobuf/ptypes/timestamp"
	"google.golang.org/api/iterator"
	monitoringpb "google.golang.org/genproto/googleapis/monitoring/v3"
)

type GCP struct {
	ctx    context.Context
	client *monitoring.MetricClient
}

// Currently needs a service account configured or gcloud auth within same session
func NewGCP() (*GCP, error) {

	ctx := context.Background()
	client, err := monitoring.NewMetricClient(ctx)
	fmt.Print(*client)
	if err != nil {
		return nil, fmt.Errorf("NewMetricClient: %v", err)
	}

	return &GCP{
		ctx:    ctx,
		client: client,
	}, nil
}

func (p *GCP) Get99PercentLatencyMetricForResource(resourceID string, from, to time.Time) (float64, error) {
	projectID := strings.SplitN(resourceID, "/", -1)[1]

	startTime := from
	endTime := to
	metricType := "loadbalancing.googleapis.com/https/total_latencies"
	// resourceName := strings.SplitN(resourceID, "/", -1)[3]

	req := &monitoringpb.ListTimeSeriesRequest{
		Name: "projects/" + projectID,
		// Filter: fmt.Sprintf(`metric.type="%s" AND resource.url_map_name="%s"`, metricType, resourceName),
		Filter: fmt.Sprintf(`metric.type="%s"`, metricType),
		Interval: &monitoringpb.TimeInterval{
			StartTime: &timestamp.Timestamp{
				Seconds: startTime.Unix(),
			},
			EndTime: &timestamp.Timestamp{
				Seconds: endTime.Unix(),
			},
		},
		Aggregation: &monitoringpb.Aggregation{
			CrossSeriesReducer: monitoringpb.Aggregation_REDUCE_MEAN,
			PerSeriesAligner:   monitoringpb.Aggregation_ALIGN_PERCENTILE_99,
			AlignmentPeriod: &duration.Duration{
				Seconds: 600,
			},
		},
	}
	it := p.client.ListTimeSeries(p.ctx, req)
	for {
		resp, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return 1, fmt.Errorf("could not read time series value: %v", err)
		}
		return resp.GetPoints()[0].GetValue().GetDoubleValue(), nil
	}

	return -1, errors.New("Get99PercentLatencyMetricForResource not implemented")
}

func (p *GCP) GetErrorPercentageMetricForResource(resourceID string, from, to time.Time) (float64, error) {
	fmt.Printf("GetErrorPercentageMetricForResource \n")
	projectID := strings.SplitN(resourceID, "/", -1)[1]

	startTime := from
	endTime := to
	metricType := "loadbalancing.googleapis.com/https/request_count"
	// resourceName := "473344846455" //strings.SplitN(resourceID, "/", -1)[3]

	req := &monitoringpb.ListTimeSeriesRequest{
		Name: "projects/" + projectID,
		// Filter: fmt.Sprintf(`metric.type = "%s" AND resource.url_map_name = "%s" AND
		// metric.response_code_class != 0`, metricType, resourceName),
		// Filter: fmt.Sprintf(`metric.type = "%s" AND
		// metric.response_code_class != 0`, metricType),
		Filter: fmt.Sprintf(`metric.type="%s"`, metricType),
		//resource.project_id == '473344846455'
		//&& (resource.url_map_name == 'reliablyadvicealpha1')
		//&& (metric.response_code_class != 0)
		// Filter: fmt.Sprintf(`metric.type ="%s"`, metricType),
		Interval: &monitoringpb.TimeInterval{
			StartTime: &timestamp.Timestamp{
				Seconds: startTime.Unix(),
			},
			EndTime: &timestamp.Timestamp{
				Seconds: endTime.Unix(),
			},
		},
		Aggregation: &monitoringpb.Aggregation{
			CrossSeriesReducer: monitoringpb.Aggregation_REDUCE_SUM,
			PerSeriesAligner:   monitoringpb.Aggregation_ALIGN_COUNT,
			AlignmentPeriod: &duration.Duration{
				Seconds: 600,
			},
			// Aggregation: &monitoringpb.Aggregation{
			// 	CrossSeriesReducer: monitoringpb.Aggregation_REDUCE_MEAN,
			// 	PerSeriesAligner:   monitoringpb.Aggregation_ALIGN_PERCENTILE_99,
			// 	AlignmentPeriod: &duration.Duration{
			// 		Seconds: 600,
			// 	},
		},
	}
	fmt.Println("Found data points for the following instances:")
	it := p.client.ListTimeSeries(p.ctx, req)
	for {
		resp, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return 1, fmt.Errorf("could not read time series value: %v", err)
		}
		// fmt.Printf("Metric: %v\n", resp)
		fmt.Printf("Metric: %v\n", resp.GetPoints()[0].GetValue().GetInt64Value())
	}
	fmt.Println("Done")
	return 1, nil

	// return -1, errors.New("GetErrorPercentageMetricForResource not implemented")
}
