package history

import (
	//"errors"
	"fmt"
	//"math/rand"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/go-echarts/go-echarts/v2/types"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/reliablyhq/cli/api"
	"github.com/reliablyhq/cli/cmd/reliably/cmdutil"
	"github.com/reliablyhq/cli/core"
	"github.com/reliablyhq/cli/core/iostreams"
	"github.com/reliablyhq/cli/core/report"
	"github.com/reliablyhq/cli/utils"
)

type HistoryOptions struct {
	IO         *iostreams.IOStreams
	HttpClient func() *http.Client
	Hostname   string

	OrgID string

	LimitResults int
	Web          bool
}

func NewCmdHistory() *cobra.Command {
	opts := &HistoryOptions{
		IO: iostreams.System(),
		HttpClient: func() *http.Client {
			return api.AuthHTTPClient(core.Hostname())
		},
	}

	cmd := &cobra.Command{
		Use:   "history",
		Short: "Show your SLO history",
		Long:  `Show the evolution of your SLOs over time.`,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			if !cmdutil.CheckAuth() {
				cmdutil.PrintRequireAuthMsg()
				os.Exit(1)
			}
		},
		PreRunE: func(cmd *cobra.Command, args []string) error {
			var err error
			apiClient := api.NewClientFromHTTP(opts.HttpClient())

			// Ensure the CLI history is executed in a valid org
			opts.OrgID, err = api.CurrentUserOrganizationID(apiClient, core.Hostname())
			if err != nil {
				return fmt.Errorf("unable to retrieve current organization: %w", err)
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Hostname = core.Hostname()

			if opts.LimitResults < 1 {
				return fmt.Errorf("invalid value for --limit: %v", opts.LimitResults)
			}

			return historyRun(opts)
		},
	}

	cmd.Flags().BoolVar(&opts.Web, "web", false, "Display history of SLO reports in your web browser")
	cmd.Flags().IntVarP(&opts.LimitResults, "limit", "l", 10, "Maximum number of reports to fetch")

	return cmd
}

func historyRun(opts *HistoryOptions) (err error) {

	opts.IO.StartProgressIndicator()

	apiClient := api.NewClientFromHTTP(opts.HttpClient())
	reports, err := api.GetReports(apiClient, opts.Hostname, opts.OrgID, opts.LimitResults)

	opts.IO.StopProgressIndicator()

	if err != nil {
		return fmt.Errorf("Unable to retrieve your history of SLO reports: %w", err)
	}

	if opts.Web {
		_ = openHistoryInWebBrowser(reports)
	} else {
		fmt.Printf("We fetched %d reports\n", len(reports))
		for i, r := range reports {
			fmt.Printf("Report #%d - %s\n", i+1, r.Timestamp.Format(time.RFC1123))
			report.Write(report.SimpleText, &r, opts.IO.Out, log.StandardLogger())
			//fmt.Println(r)
			fmt.Printf("---\n\n")
		}
	}

	return nil
}

// localServerFlow opens the authentication page for a provider
// in a browser tab, then returns the authorization state & code
func openHistoryInWebBrowser(reports []report.Report) (err error) {
	listener, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		return
	}
	port := listener.Addr().(*net.TCPAddr).Port
	localhost := "127.0.0.1"
	localURL := fmt.Sprintf("http://%s:%d", localhost, port)

	//http.HandleFunc("/", httpserver)
	//http.ListenAndServe(":8081", nil)

	log.Debugf("open %s\n", localURL)
	err = utils.OpenInBrowser(localURL)
	if err != nil {
		return
	}

	//httpserver(w http.ResponseWriter, _ *http.Request)
	_ = http.Serve(listener, http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {

		//ll := len(reports)

		var r report.Report = reports[0]

		globalSloAreMet := make([][]bool, 0)

		for i, svc := range r.Services {

			svcDates := make([]string, 0)
			for ridx := len(reports) - 1; ridx >= 0; ridx-- {
				svcDates = append(svcDates, reports[ridx].Timestamp.String()[:19])
			}

			svcSloValues := make(map[string][]opts.LineData, 0)

			svcSloMet := make(map[string][]bool, 0)

			//svc.Name

			for j, sl := range svc.ServiceLevels {

				//sl.Name

				sloMet := make([]bool, 0)
				sloValues := make([]float64, 0)
				sloObjectives := make([]float64, 0)
				//sloDates := make([]time.Time, 0)
				sloDates := make([]string, 0)

				// get values for each report in reversed order - oldest to newest
				for ridx := len(reports) - 1; ridx >= 0; ridx-- {
					if res := reports[ridx].Services[i].ServiceLevels[j].Result; res != nil {
						sloMet = append(sloMet, res.SloIsMet)
						sloValues = append(sloValues, res.Actual.(float64))
						sloObjectives = append(sloObjectives, reports[ridx].Services[i].ServiceLevels[j].Objective)
						//sloDates = append(sloDates, reports[ridx].Timestamp)
						sloDates = append(sloDates, reports[ridx].Timestamp.String()[:19])
					}
				}

				fmt.Println("sloMet", sloMet)
				fmt.Println("slo values", sloValues)

				items := make([]opts.LineData, 0)
				for i := 0; i < len(sloValues); i++ {
					items = append(items, opts.LineData{Value: sloValues[i]})
				}

				svcSloValues[sl.Name] = items

				/*
					hmitems := make([]opts.HeatMapData, 0)
					for i := 0; i < len(sloMet); i++ {
						if sloMet[i] {
							hmitems = append(hmitems, opts.HeatMapData{Value: [1]interface{}{1}})
						} else {
							hmitems = append(hmitems, opts.HeatMapData{Value: [1]interface{}{0}})
						}

						//hmitems = append(hmitems, opts.HeatMapData{Value: sloMet[i]})
					}
				*/
				svcSloMet[sl.Name] = sloMet

				globalSloAreMet = append(globalSloAreMet, sloMet)

				thresholds := make([]opts.LineData, 0)
				for i := 0; i < len(sloObjectives); i++ {
					thresholds = append(thresholds, opts.LineData{Value: sloObjectives[i]})
				}

				fmt.Println("items", items)
				fmt.Println("thresholds", thresholds)

				line := charts.NewLine()
				// set some global options like Title/Legend/ToolTip or anything else
				line.SetGlobalOptions(
					charts.WithInitializationOpts(opts.Initialization{Theme: types.ThemeWesteros}),
					charts.WithTitleOpts(opts.Title{
						Title:    svc.Name,
						Subtitle: sl.Name,
					}))

				// Put data into instance
				//line.SetXAxis([]string{"Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"}).
				line.
					SetXAxis(sloDates).
					//SetXAxis(make([]string, len(sloValues))).
					AddSeries("SLO % ", items).
					//AddSeries("Objective", thresholds).
					SetGlobalOptions(
						charts.WithYAxisOpts(opts.YAxis{Min: 0, Max: 100}),
						charts.WithXAxisOpts(opts.XAxis{AxisLabel: &opts.AxisLabel{Rotate: 25}}),
					).
					SetSeriesOptions(
						charts.WithLineChartOpts(opts.LineChart{Smooth: false}),
						//charts.WithLabelOpts(opts.Label{Show: true}),
						/*
							charts.WithMarkPointStyleOpts(opts.MarkPointStyle{
								Label: &opts.Label{
									Show:      true,
									Formatter: "{a}: {b}",
								},
							}),
						*/
						charts.WithMarkLineNameYAxisItemOpts(opts.MarkLineNameYAxisItem{
							Name:  "Objective",
							YAxis: sl.Objective,
						}),
					)

				line.Render(w)

			}

			// chart for service - all SLOs are on the same chart

			line := charts.NewLine()
			// set some global options like Title/Legend/ToolTip or anything else
			line.SetGlobalOptions(
				charts.WithInitializationOpts(opts.Initialization{Theme: types.ThemeWesteros}),
				charts.WithTitleOpts(opts.Title{
					Title:    svc.Name,
					Subtitle: "All SLOs",
				}))

			line.
				SetXAxis(svcDates).
				//SetXAxis(make([]string, len(sloValues))).

				//AddSeries("Objective", thresholds).
				SetGlobalOptions(
					charts.WithYAxisOpts(opts.YAxis{Min: 0, Max: 100}),
					charts.WithXAxisOpts(opts.XAxis{AxisLabel: &opts.AxisLabel{Rotate: 25}}),
				).
				SetSeriesOptions(
					charts.WithLineChartOpts(opts.LineChart{Smooth: false}),
					//charts.WithLabelOpts(opts.Label{Show: true}),
					/*
						charts.WithMarkPointStyleOpts(opts.MarkPointStyle{
							Label: &opts.Label{
								Show:      true,
								Formatter: "{a}: {b}",
							},
						}),
					*/
					/*
						charts.WithMarkLineNameYAxisItemOpts(opts.MarkLineNameYAxisItem{
							Name:  "Objective",
							YAxis: sl.Objective,
						}),
					*/
				)

			for sloName, sloItems := range svcSloValues {
				line.AddSeries(sloName, sloItems)
			}
			line.Render(w)

			// HEATMAP for SLO per service

			sloNames := make([]string, 0)
			for k := range svcSloMet {
				sloNames = append(sloNames, k)
			}

			fmt.Println("slo names", sloNames)

			xaxis := [...]string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10"}

			/*
				hm := charts.NewHeatMap()
				hm.SetGlobalOptions(
					charts.WithTitleOpts(opts.Title{
						Title: fmt.Sprintf("heatmap of SLOs for service %s", svc.Name),
					}),
					charts.WithXAxisOpts(opts.XAxis{
						Type:      "dates",
						Data:      xaxis,
						SplitArea: &opts.SplitArea{Show: true},
					}),
					charts.WithYAxisOpts(opts.YAxis{
						Type:      "SLOs",
						Data:      sloNames,
						SplitArea: &opts.SplitArea{Show: true},
					}),
					charts.WithVisualMapOpts(opts.VisualMap{
						Calculable: false,
						Min:        0,
						Max:        1,
						InRange: &opts.VisualMapInRange{
							Color: []string{"#50a3ba", "#eac736"},
						},
					}),
				)
			*/

			hm := charts.NewHeatMap()
			hm.SetGlobalOptions(
				charts.WithTitleOpts(opts.Title{
					Title: fmt.Sprintf("heatmap of SLOs for service %s", svc.Name),
				}),
				charts.WithXAxisOpts(opts.XAxis{
					Type:      "category",
					Data:      xaxis,
					SplitArea: &opts.SplitArea{Show: true},
				}),
				charts.WithYAxisOpts(opts.YAxis{
					Type:      "category",
					Data:      sloNames,
					SplitArea: &opts.SplitArea{Show: true},
				}),
				charts.WithVisualMapOpts(opts.VisualMap{
					Calculable: true,
					Min:        0,
					Max:        1,
					InRange: &opts.VisualMapInRange{
						Color: []string{"#7db584", "#d94e5d"},
					},
				}),
			)

			//hm.SetXAxis(dayHrs).AddSeries("heatmap", genHeatMapData())

			//hm.SetXAxis([...]string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10"})

			//allSvcSloMet := make([]opts.HeatMapData, 0)

			hmitems := make([]opts.HeatMapData, 0)

			for ii, name := range sloNames {
				met := svcSloMet[name]

				for jj, b := range met {
					var v int = 0
					if b {
						v = 1
					}

					hmitems = append(hmitems, opts.HeatMapData{Value: [3]interface{}{jj, ii, v}})
				}
				//allSvcSloMet

			}

			/*

				for i := 0; i < len(sloMet); i++ {
					if sloMet[i] {
						hmitems = append(hmitems, opts.HeatMapData{Value: [1]interface{}{1}})
					} else {
						hmitems = append(hmitems, opts.HeatMapData{Value: [1]interface{}{0}})
					}

					//hmitems = append(hmitems, opts.HeatMapData{Value: sloMet[i]})
				}
			*/

			/*
				for _, sloMet := range svcSloMet {
					allSvcSloMet = append(allSvcSloMet, sloMet...)
					//fmt.Println("HEATMAP values for", sloName, sloMet)

					//line.AddSeries(sloName, sloItems)
				}
			*/

			//hm.AddSeries("heatmap", allSvcSloMet)
			//fmt.Println("HEATMAP all values", allSvcSloMet)
			hm.AddSeries("heatmap", hmitems)

			_ = hm.Render(w)

			//hm.Validate()

		}

		// GLOBAL HEATMAP, for all SLOs, for all Services

		xaxis := [...]string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10"}

		hmitems := make([]opts.HeatMapData, 0)
		hm := charts.NewHeatMap()
		hm.SetGlobalOptions(
			charts.WithTitleOpts(opts.Title{
				Title: "heatmap of all SLOs",
			}),
			charts.WithXAxisOpts(opts.XAxis{
				Name:      "reports",
				Type:      "category",
				Data:      xaxis,
				SplitArea: &opts.SplitArea{Show: true},
			}),
			charts.WithYAxisOpts(opts.YAxis{
				Name: "SLO",
				Type: "category",
				//Data:      sloNames,
				SplitArea: &opts.SplitArea{Show: true},
			}),
			charts.WithVisualMapOpts(opts.VisualMap{
				Calculable: true,
				Min:        0,
				Max:        1,
				InRange: &opts.VisualMapInRange{
					Color: []string{"#7db584", "#d94e5d"},
				},
			}),
		)

		//hm.SetXAxis(dayHrs).AddSeries("heatmap", genHeatMapData())

		//hm.SetXAxis([...]string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10"})

		//allSvcSloMet := make([]opts.HeatMapData, 0)

		for ii, met := range globalSloAreMet {
			//met := svcSloMet[name]

			if len(met) == 0 {
				for jj := 0; jj < 10; jj++ {
					hmitems = append(hmitems, opts.HeatMapData{Value: [3]interface{}{jj, ii, "-"}})
				}
			}

			for jj, b := range met {
				var v int = 0
				if b {
					v = 1
				}

				hmitems = append(hmitems, opts.HeatMapData{Value: [3]interface{}{jj, ii, v}})
			}
			//allSvcSloMet

		}

		/*

			for i := 0; i < len(sloMet); i++ {
				if sloMet[i] {
					hmitems = append(hmitems, opts.HeatMapData{Value: [1]interface{}{1}})
				} else {
					hmitems = append(hmitems, opts.HeatMapData{Value: [1]interface{}{0}})
				}

				//hmitems = append(hmitems, opts.HeatMapData{Value: sloMet[i]})
			}
		*/

		/*
			for _, sloMet := range svcSloMet {
				allSvcSloMet = append(allSvcSloMet, sloMet...)
				//fmt.Println("HEATMAP values for", sloName, sloMet)

				//line.AddSeries(sloName, sloItems)
			}
		*/

		//fmt.Println("LEN", len(xaxis), len(sloNames), len(hmitems))

		//hm.AddSeries("heatmap", allSvcSloMet)
		//fmt.Println("HEATMAP all values", allSvcSloMet)
		hm.AddSeries("heatmap", hmitems)
		fmt.Println("HEATMAP all values", len(hmitems), hmitems)

		fmt.Println()
		err := hm.Render(w)
		if err != nil {
			fmt.Println("ERRROR for heatmap", err)
		}

		//httpserver(w, req)
		//defer listener.Close()
	}))
	/*
		_ = http.Serve(listener, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log.Debugf("server handler: %s\n", r.URL.Path)
			if r.URL.Path != callbackPath {
				w.WriteHeader(404)
				return
			}
			defer listener.Close()
			rq := r.URL.Query()
			if state != rq.Get("state") {
				fmt.Fprintf(w, "Error: state mismatch")
				return
			}
			log.Debugf("server received query params %s", rq)
			code = rq.Get("code")
			log.Debugf("server received code %q\n", code)

			w.Header().Add("content-type", "text/html")
			//fmt.Fprintf(w, "<p>You have successfully authenticated. You may now close this page.</p>")
			fmt.Fprint(w, oauthSuccessPage)


		}))
	*/

	return
}

/*
// generate random data for line chart
func generateLineItems() []opts.LineData {
	items := make([]opts.LineData, 0)
	for i := 0; i < 7; i++ {
		items = append(items, opts.LineData{Value: rand.Intn(300)})
	}
	return items
}
*/

/*
func httpserver(w http.ResponseWriter, _ *http.Request) {
	// create a new line instance
	line := charts.NewLine()
	// set some global options like Title/Legend/ToolTip or anything else
	line.SetGlobalOptions(
		charts.WithInitializationOpts(opts.Initialization{Theme: types.ThemeWesteros}),
		charts.WithTitleOpts(opts.Title{
			Title:    "Line example in Westeros theme",
			Subtitle: "Line chart rendered by the http server this time",
		}))

	// Put data into instance
	line.SetXAxis([]string{"Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"}).
		AddSeries("Category A", generateLineItems()).
		AddSeries("Category B", generateLineItems()).
		SetSeriesOptions(charts.WithLineChartOpts(opts.LineChart{Smooth: true}))
	line.Render(w)

	line2 := charts.NewLine()
	// set some global options like Title/Legend/ToolTip or anything else
	line2.SetGlobalOptions(
		charts.WithInitializationOpts(opts.Initialization{Theme: types.ThemeWesteros}),
		charts.WithTitleOpts(opts.Title{
			Title:    "Line example in Westeros theme",
			Subtitle: "Line chart rendered by the http server this time",
		}))

	// Put data into instance
	line2.SetXAxis([]string{"Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"}).
		AddSeries("Category A", generateLineItems()).
		AddSeries("Category B", generateLineItems()).
		SetSeriesOptions(charts.WithLineChartOpts(opts.LineChart{Smooth: true}))
	line2.Render(w)

}

*/
