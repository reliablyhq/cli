package init

import (
	"fmt"
	"strings"
	"time"

	iso8601 "github.com/ChannelMeter/iso8601duration"
	"github.com/davecgh/go-spew/spew"
	"github.com/reliablyhq/cli/core"

	"github.com/reliablyhq/cli/core/manifest"
	"github.com/reliablyhq/cli/core/metrics/datadog"
)

func QueryMetrics(query string) error {
	pl, err := datadog.RunQueryMetrics(query)
	fmt.Println(pl)
	return err
}

func ImportSLOsFromDatadog() error {

	fmt.Println("Validate API Key for datadog")
	if ok, err := datadog.ValidateApiKey(); ok {
		fmt.Println("Authenticated to Datadog API")
	} else {
		fmt.Println("Error while validating DD API KEY", err)
	}

	var slosDD []*manifest.ServiceLevel = make([]*manifest.ServiceLevel, 0)

	fmt.Println("List SLOs from datadog")
	if r, err := datadog.ListDatadogSLOs(); err != nil {
		fmt.Println("Error with datadog", err)
	} else {
		fmt.Println("we found your SLOs")

		fmt.Println("-> errors", r.Errors)
		fmt.Println("-> metadata", r.Metadata.Page.GetTotalFilteredCount(), r.Metadata.Page.GetTotalCount())
		fmt.Println("-> data", r.GetData())
		for _, slo := range r.GetData() {

			for _, target := range slo.Thresholds {

				n := fmt.Sprintf("%s - %s ", slo.Name, target.Timeframe)
				fmt.Println("Name, Target, Timeframe")
				fmt.Println(n, target.Target, target.Timeframe)

				dur, _ := iso8601.FromString(strings.ToUpper(fmt.Sprintf("P%s", target.Timeframe)))

				slosDD = append(slosDD, &manifest.ServiceLevel{
					Name:      n,
					Type:      "mirror",
					Objective: target.Target,
					Indicators: []manifest.ServiceLevelIndicator{
						{
							ID:       *slo.Id,
							Provider: "datadog",
						},
					},
					ObservationWindow: core.Iso8601Duration{
						Duration: *dur,
					},
				})

			}
			spew.Dump(slo)
		}
	}

	var mDD manifest.Manifest = manifest.Manifest{

		Services: []*manifest.Service{
			{
				Name:          "datadog imported SLOs",
				ServiceLevels: slosDD,
			},
		},
	}

	fmt.Println("Manifest with SLOs imported from datadog")
	spew.Dump(mDD)

	fmt.Println("Get SLO history from datadog")
	for _, svc := range mDD.Services {
		fmt.Println("###", svc.Name)
		for _, slo := range svc.ServiceLevels {
			fmt.Println(">>>", slo.Name, slo.Objective, slo.ObservationWindow)
			sli := slo.Indicators[0]

			to := time.Now()

			from := to.Add(-slo.ObservationWindow.ToDuration())

			//from := to.Add(-time.Hour * 24 * 7)

			dur := to.Sub(from).Hours()

			fmt.Println("GetSLOHistory (sloId, from, to, target) - duration in hours")
			fmt.Println(sli.ID, from, to, slo.Objective, dur, from.UTC().Unix(), to.UTC().Unix())

			if sloHist, err := datadog.GetSLOHistory(sli.ID, from, to, slo.Objective); err != nil {
				fmt.Println("Unable to fetch SLO history", err)
			} else {
				fmt.Println("we found the history -->")
				data := sloHist.GetData()
				var sliValue float64
				if data.Overall.SliValue != nil {
					sliValue = *data.Overall.SliValue
					fmt.Println("SLI", "=", sliValue, "%")

					if data.Overall.ErrorBudgetRemaining != nil {
						for k, v := range *data.Overall.ErrorBudgetRemaining {
							fmt.Println("Error budget for ", k, v, "%")
						}
					} else {
						fmt.Println("error budget is not available")
					}

				} else {
					fmt.Println("No SLI value computed ! ")
				}

				//spew.Dump(data)
			}

		}
	}

	return fmt.Errorf("Skip")
}
