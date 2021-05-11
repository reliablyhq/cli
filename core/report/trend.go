package report

// SLOTrend returns the trend for a given SLO is all reports
// for each report, we need to iterate over any service/slo as indexes
// might evolve over time, from one report to another
// It returns a slice of boolean values when SLO is met or not for each report
// When no result is available from a given report, no trend is appended
func GetSLOTrend(svcName string, sloName string, reports []Report) []bool {

	var trend []bool = make([]bool, 0)

	for _, r := range reports {
		res := r.GetResult(svcName, sloName)
		if res != nil {
			trend = append(trend, res.SloIsMet)
		}
	}

	return trend
}
