package datadog

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"

	datadog "github.com/DataDog/datadog-api-client-go/api/v1/datadog"
	log "github.com/sirupsen/logrus"

	"github.com/reliablyhq/cli/utils"
)

type Pointlist *[][]float64

func RunQueryMetrics(query string) (Pointlist, error) {

	/*
		to, _ := time.Parse(time.RFC3339, "2021-07-13T00:00:00Z")
		from := to.Add(-time.Hour * (24 * 7))
	*/

	now := time.Now().Unix()
	to := now
	week := time.Hour * 24 * 7
	from := now - int64(week.Seconds())

	log.Debug("Run datadog metrics query")
	log.Debugf("from: %s - to: %s", from, to)
	log.Debug("query=", query)

	ctx := datadog.NewDefaultContext(context.Background())

	configuration := datadog.NewConfiguration()
	apiClient := datadog.NewAPIClient(configuration)

	fmt.Println("query")
	fmt.Println(query)

	/*
		now := time.Now().Unix()
		to := now
		week := time.Hour * 24 * 7
		from := now - int64(week.Seconds())
	*/

	queryResult, _, err := apiClient.MetricsApi.QueryMetrics(ctx, from, to, query)
	if err != nil {
		return nil, err
	}

	fmt.Println("response")
	fmt.Println(queryResult)

	fmt.Println(queryResult.GetGroupBy())
	fmt.Println(queryResult.GetQuery())
	fmt.Println(queryResult.GetFromDate())
	fmt.Println(queryResult.GetToDate())
	fmt.Println(queryResult.GetStatus())
	fmt.Println(queryResult.GetResType())
	fmt.Println(len(queryResult.GetSeries()))

	queryResult.GetSeriesOk()
	if len(queryResult.GetSeries()) == 0 {
		return nil, errors.New("No data series retrieved from metrics")
	}

	series := queryResult.GetSeries()[0]

	sum := 0.0
	datapoints := series.GetPointlist()
	for _, dp := range datapoints {

		sec, dec := math.Modf(dp[0] / 1000)
		//fmt.Println(dp[0], sec, dec)
		t := time.Unix(int64(sec), int64(dec*(1e9)))

		fmt.Println("> ", t, dp[1])

		sum = sum + dp[1]

	}
	fmt.Println("SUM", sum)
	fmt.Println("AVG", sum/float64(len(datapoints)))
	/*
		if len(datapoints) == 0 {
			return errors.New("No data points retrieved from metrics series")
		}

		dp := datapoints[0]
		for _

	*/
	fmt.Println(series.GetLength())
	fmt.Println(series.GetAggr())
	fmt.Println(series.GetDisplayName())
	fmt.Println(series.GetMetric())
	fmt.Println(series.GetPointlist())
	/*
		series.GetPointlist()[0][0], float64(series.GetStart()))
		series.GetPointlist()[1][0], float64(series.GetEnd()))
		10.5, series.GetPointlist()[0][1])
		11., series.GetPointlist()[1][1])
	*/

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

func ComputeSloFromQueryMetrics(numerator_query string, denominator_query string) (float64, error) {

	numResult, err := RunQueryMetrics(numerator_query)
	if err != nil {
		return 0.0, err
	}

	denomResult, err := RunQueryMetrics(denominator_query)
	if err != nil {
		return 0.0, err
	}

	numMap := pointlist2Map(numResult)
	denomMap := pointlist2Map(denomResult)

	slo := computeSLOWithNumDenom(numMap, denomMap)

	fmt.Println("numerator map", numMap)
	fmt.Println("denominator map", denomMap)

	return slo, nil

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
