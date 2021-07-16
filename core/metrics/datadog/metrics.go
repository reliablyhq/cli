package datadog

import (
	"context"
	"errors"
	"time"

	datadog "github.com/DataDog/datadog-api-client-go/api/v1/datadog"
	log "github.com/sirupsen/logrus"

	"github.com/reliablyhq/cli/utils"
)

type Pointlist *[][]float64

func RunQueryMetrics(query string, from time.Time, to time.Time) (Pointlist, error) {

	log.Debug("Run datadog metrics query")
	log.Debugf("from: %s - to: %s", from, to)
	log.Debug("query=", query)

	ctx := datadog.NewDefaultContext(context.Background())

	configuration := datadog.NewConfiguration()
	apiClient := datadog.NewAPIClient(configuration)

	queryResult, _, err := apiClient.MetricsApi.QueryMetrics(ctx, from.Unix(), to.Unix(), query)
	if err != nil {
		return nil, err
	}

	queryResult.GetSeriesOk()
	if len(queryResult.GetSeries()) == 0 {
		return nil, errors.New("No data series retrieved from metrics")
	}

	series := queryResult.GetSeries()[0]
	datapoints := series.GetPointlist()
	return &datapoints, err
}

// pointlist2Map converts a list of datapoints [(x, y), ...] into
// a map where key is the datapoint x value and key the y value
func pointlist2Map(pl Pointlist) map[float64]float64 {

	var m map[float64]float64 = make(map[float64]float64)

	if pl == nil {
		return m
	}

	for _, dp := range *pl {
		x, y := dp[0], dp[1]
		m[x] = y
	}

	return m

}

func ComputeSloFromQueryMetrics(numerator_query string, denominator_query string, from time.Time, to time.Time) (float64, error) {

	numResult, err := RunQueryMetrics(numerator_query, from, to)
	if err != nil {
		return 0.0, err
	}

	denomResult, err := RunQueryMetrics(denominator_query, from, to)
	if err != nil {
		return 0.0, err
	}

	numMap := pointlist2Map(numResult)
	denomMap := pointlist2Map(denomResult)

	slo := computeSLOWithNumDenom(numMap, denomMap)

	return slo * 100.0, nil

}

// computeSLOWithNumDenom computes the averaged SLO value from all
// numerator over denominator datapoints
func computeSLOWithNumDenom(num, denom map[float64]float64) float64 {

	var values []float64
	for ts, d := range denom {
		n := 0.0
		if val, ok := num[ts]; ok {
			n = float64(val)
		}
		values = append(values, float64(n/d))
	}

	return utils.AvgFloat64(values)
}
