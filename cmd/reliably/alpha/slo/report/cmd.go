package reportAlpha

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"sort"
	"syscall"
	"time"

	iso8601 "github.com/ChannelMeter/iso8601duration"
	"github.com/reliablyhq/cli/api"
	sloReport "github.com/reliablyhq/cli/cmd/reliably/slo/report"
	"github.com/reliablyhq/cli/config"
	"github.com/reliablyhq/cli/core"
	"github.com/reliablyhq/cli/core/color"
	"github.com/reliablyhq/cli/core/entities"
	"github.com/reliablyhq/cli/core/report"
	"github.com/reliablyhq/cli/utils"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

func AlpaReportRun(opts *sloReport.ReportOptions) error {

	// check for -w/--watch
	if opts.WatchFlag {
		return watch(opts)
	}

	opts.IO.StartProgressIndicator()

	reports, err := getReports(opts.ManifestPath)
	if err != nil {
		return fmt.Errorf("reports error: %w", err)
	}

	opts.IO.StopProgressIndicator()

	for _, out := range opts.Outputs {

		var w io.Writer = os.Stdout
		if out.Path != "" {
			outfile, err := os.Create(out.Path) // creates or truncates with O_RDWR mode
			if err != nil {
				log.Error("error creating output file")
				log.Error(err)
				return err
			}
			w = outfile
			// we cannot defer outfile closing here as we are in a for-loop
		}
		// utils.Reverse(reportCollection)
		report.Write(out.Format, reports[0], w, opts.TemplateFile, log.StandardLogger(), reports[1], editReportSlice(reports))

		if outfile, ok := w.(*os.File); ok {
			outfile.Close() // explicitly closing the file handle
		}

	}

	return nil
}

func getReports(manifestPath string) ([]*report.Report, error) {
	apiVersion := "v1"
	reportsLimit := 5

	objectives, err := loadObjectives(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read manifest: %w", err)
	}

	hostname := config.Hostname
	entityHost := config.EntityServerHost

	apiClient := api.NewClientFromHTTP(api.AuthHTTPClient(hostname))
	org, err := api.CurrentUserOrganization(apiClient, hostname)
	if err != nil {
		return nil, fmt.Errorf("unable to request org: %w", err)
	}

	// TODO: define version/kind from manifest objective?
	objectiveResults, err := api.GetObjectiveResults(apiClient, entityHost, apiVersion, org.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to get objective results: %w", err)
	}

	filteredObjectiveResults := filterObjectivesResults(objectiveResults, objectives, reportsLimit)

	// Important: at the moment each objective result represents the difference between
	//  the objective and the indicator at that time. If the objective is updated, only
	// indicators after it will produce an objective result with the delta of the new one. An alternative is to always
	// Use the latest objective against objective results. So, either an objective change updates previous objective results
	// retrospectively, or each stands on its own.
	reports, err := MapToReports(filteredObjectiveResults, reportsLimit, apiVersion)
	if err != nil {
		return nil, fmt.Errorf("failed to generate report: %w", err)
	}

	return reports, nil
}

func editReportSlice(s []*report.Report) *[]report.Report {
	var sNew []report.Report
	for _, v := range s {
		sNew = append(sNew, *v)
	}
	utils.Reverse(sNew)
	return &sNew
}

// Function filters objectiveResultResponses based on objectives.
// It is assumed the slice is ordered in descending order by creation time.
// If Entity Server assumes filtering, this function can likely be replaced.
func filterObjectivesResults(
	objectiveResults *[]entities.ObjectiveResultResponse,
	objectives []entities.Objective,
	maxResults int,
) [][]entities.ObjectiveResultResponse {

	// Creating a hash table for faster search, although likely not many entities.
	// Array of name+service used as map key.
	objectivesMapped := make(map[[2]string]int)
	for i, o := range objectives {
		if _, ok := o.Metadata.Labels["name"]; !ok {
			continue
		}
		if _, ok := o.Metadata.Labels["service"]; !ok {
			continue
		}
		objectivesMapped[[2]string{
			o.Metadata.Labels["name"], o.Metadata.Labels["service"],
		}] = i
	}

	filteredObjRes := make([][]entities.ObjectiveResultResponse, len(objectives))
	for _, or := range *objectiveResults {
		if _, ok := or.Metadata.Labels["name"]; !ok {
			continue
		}
		if _, ok := or.Metadata.Labels["service"]; !ok {
			continue
		}
		nameAndService := [2]string{or.Metadata.Labels["name"], or.Metadata.Labels["service"]}
		if objectiveIndex, ok := objectivesMapped[nameAndService]; ok {
			if len(filteredObjRes[objectiveIndex]) <= maxResults {
				filteredObjRes[objectiveIndex] = append(filteredObjRes[objectiveIndex], or)
			}
		}
	}

	filteredObjResClean := make([][]entities.ObjectiveResultResponse, 0)
	for _, v := range filteredObjRes {
		if v != nil {
			filteredObjResClean = append(filteredObjResClean, v)
		}
	}

	return filteredObjResClean
}

func loadObjectives(path string) ([]entities.Objective, error) {
	var objects []entities.Objective = make([]entities.Objective, 0)

	if path == "" {
		return nil, errors.New("path is empty")
	}

	log.Debug("Loading manifest at ", path)

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	//var m Manifest
	dec := yaml.NewDecoder(file)

	var objective *entities.Objective
	for dec.Decode(&objective) == nil {
		objects = append(objects, *objective)
		// ensure to create a new pointer for next iteration - avoid merged sub-props
		objective = new(entities.Objective)
	}

	return objects, nil
}

func MapToReports(objResults [][]entities.ObjectiveResultResponse, limit int, apiVersion string) ([]*report.Report, error) {
	var mappedReports []*report.Report

	for i := 0; i < limit; i++ {
		var services []*report.Service = make([]*report.Service, 0)
		mappedReports = append(mappedReports, &report.Report{
			APIVersion: apiVersion,
			Timestamp:  time.Now().UTC(),
			Services:   services,
		})

		serviceList := make(map[string]map[string][]entities.ObjectiveResultResponse)
		for _, objResGroup := range objResults {
			serviceLabel := objResGroup[0].Metadata.Labels["service"]
			nameLabel := objResGroup[0].Metadata.Labels["name"]
			if _, ok := serviceList[serviceLabel]; !ok {
				serviceList[serviceLabel] = make(map[string][]entities.ObjectiveResultResponse)
			}
			serviceList[serviceLabel][nameLabel] = objResGroup
		}

		// 2. each service has many service levels
		for serviceLabel, s := range serviceList {
			// 1. Define service struct
			serviceLevels := make([]*report.ServiceLevel, 0)
			service := report.Service{
				Name:          serviceLabel,
				Dependencies:  []string{},
				ServiceLevels: serviceLevels,
			}

			for name, sl := range s {
				if len(sl) > i {
					sloIsMet := false
					if sl[i].Spec.RemainingPercent >= 0 {
						sloIsMet = true
					}
					to, err := isoTimeParse(sl[i].Metadata.Labels["to"])
					if err != nil {
						return nil, fmt.Errorf("time 'to' not parsed correctly: %w", err)
					}
					from, err := isoTimeParse(sl[i].Metadata.Labels["from"])
					if err != nil {
						return nil, fmt.Errorf("time 'from' not parsed correctly: %w", err)
					}
					// Remove this once entity server returns period or calculated by from/to
					timeDiff := to.Sub(from)
					period := toIso8601Duration(timeDiff)

					service.ServiceLevels = append(service.ServiceLevels, &report.ServiceLevel{
						Name:      name,
						Type:      sl[i].Spec.IndicatorSelector["category"],
						Objective: sl[i].Spec.ObjectivePercent,
						Period:    period,
						Result: &report.ServiceLevelResult{
							Actual:   sl[i].Spec.ActualPercent,
							Delta:    sl[i].Spec.RemainingPercent,
							SloIsMet: sloIsMet,
						},
						ObservationWindow: report.Window{
							To:   to,
							From: from,
						},
					})
				}

				sort.SliceStable(service.ServiceLevels, func(k, j int) bool {
					return service.ServiceLevels[k].Name < service.ServiceLevels[j].Name
				})

			}

			mappedReports[i].Services = append(mappedReports[i].Services, &service)
		}

		sort.SliceStable(mappedReports[i].Services, func(k, j int) bool {
			return mappedReports[i].Services[k].Name < mappedReports[i].Services[j].Name
		})

	}

	return mappedReports, nil
}

func watch(opts *sloReport.ReportOptions) error {
	rChan := make(chan []*report.Report, 5)
	errChan := make(chan error, 1)
	done := make(chan struct{})
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	defer func() {
		// put cursor back on return
		fmt.Print("\033[?25h")
	}()

	// refresh every 3 seconds
	go func() {
		for ch := time.Tick(time.Second * 3); ; <-ch {
			reports, err := getReports(opts.ManifestPath)
			if err != nil {
				errChan <- err
			}
			rChan <- reports
		}
	}()

	// Ctrl+C listener
	go func() {
		<-c
		fmt.Printf("\nCTRL+C pressed... exiting\n")
		done <- struct{}{}
	}()

	// print stuff
	for {
		select {
		case r := <-rChan:
			sloReport.ClearScreen()
			fmt.Println(color.Magenta("Refreshing SLO report every 3 seconds."), "Press CTRL+C to quit.")
			report.Write(report.TABBED, r[0], os.Stdout, opts.TemplateFile, log.StandardLogger(), r[1], editReportSlice(r))

		case err := <-errChan:
			return err

		case <-done:
			return nil
		}
	}

}

// Temporary way of handling incoming time strings
func isoTimeParse(sTime string) (time.Time, error) {
	msOptions := []string{"000", "00", "0"}
	var err error
	var parsedTime time.Time
	for _, v := range msOptions {
		parsedTime, err = time.Parse(fmt.Sprintf("2006-01-02 15:04:05.%v -0700 MST", v), sTime)
		if err == nil {
			break
		}
	}
	if err != nil {
		return time.Time{}, fmt.Errorf("time not parsed correctly: %w", err)
	}
	return parsedTime, nil
}

func toIso8601Duration(d time.Duration) core.Iso8601Duration {
	di := int(d.Seconds())
	isoD := core.Iso8601Duration{Duration: iso8601.Duration{Seconds: di}}
	return isoD
}
