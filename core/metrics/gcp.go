package metrics

import (
	"context"
	"errors"
	"fmt"
	"time"

	monitoring "cloud.google.com/go/monitoring/apiv3"
	"github.com/golang/protobuf/ptypes/timestamp"
	"google.golang.org/api/iterator"
	monitoringpb "google.golang.org/genproto/googleapis/monitoring/v3"
)

type GCP struct {
	ctx    context.Context
	client *monitoring.MetricClient
}

func NewGCP() (*GCP, error) {

	ctx := context.Background()
	client, err := monitoring.NewMetricClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("NewMetricClient: %v", err)
	}

	return &GCP{
		ctx:    ctx,
		client: client,
	}, errors.New("NewGCP not implemented")
}

func (p *GCP) GetMetricList(projectID string) error {

	startTime := time.Now().UTC().Add(time.Minute * -20)
	endTime := time.Now().UTC()
	req := &monitoringpb.ListTimeSeriesRequest{
		Name:   "projects/" + projectID,
		Filter: `metric.type="compute.googleapis.com/instance/cpu/utilization"`,
		Interval: &monitoringpb.TimeInterval{
			StartTime: &timestamp.Timestamp{
				Seconds: startTime.Unix(),
			},
			EndTime: &timestamp.Timestamp{
				Seconds: endTime.Unix(),
			},
		},
		View: monitoringpb.ListTimeSeriesRequest_HEADERS,
	}
	fmt.Println("Found data points for the following instances:")
	it := p.client.ListTimeSeries(p.ctx, req)
	for {
		resp, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return fmt.Errorf("could not read time series value: %v", err)
		}
		fmt.Printf("\t%v\n", resp.GetMetric().GetLabels()["instance_name"])
	}
	fmt.Println("Done")
	return nil

}

func (p *GCP) GetLatencyMetricForResource(resourceID string, from, to time.Time) (float64, error) {
	return -1, errors.New("GetLatencyMetricForResource not implemented")
}

func (p *GCP) GetErrorPercentageMetricForResource(resourceID string, from, to time.Time) (float64, error) {
	return -1, errors.New("GetErrorPercentageMetricForResource not implemented")
}
