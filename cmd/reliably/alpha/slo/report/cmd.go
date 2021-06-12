package reportAlpha

import (
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	iso8601 "github.com/ChannelMeter/iso8601duration"
	"github.com/reliablyhq/cli/api"
	sloReport "github.com/reliablyhq/cli/cmd/reliably/slo/report"
	"github.com/reliablyhq/cli/core"
	"github.com/reliablyhq/cli/core/entities"
	"github.com/reliablyhq/cli/core/report"
	v "github.com/reliablyhq/cli/version"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

func AlpaReportRun(opts *sloReport.ReportOptions) error {
	// TODO: add more validation
	// TODO: tests required

	// TODO: implement watch for entity server
	// check for -w/--watch
	// if opts.WatchFlag {
	// 	return watch(opts)
	// }

	// TODO: define elsewhere
	apiVersion := "v1"
	reportsLimit := 6

	opts.IO.StartProgressIndicator()

	objectives, err := loadObjectives(opts.ManifestPath)
	if err != nil {
		return fmt.Errorf("failed to read manifest: %w", err)
	}
	_ = objectives

	hostname := core.Hostname()
	entityHost := core.Hostname()
	if v.IsDevVersion() {
		if hostFromEnv := os.Getenv("RELIABLY_ENTITY_HOST"); hostFromEnv != "" {
			entityHost = hostFromEnv
		}
	}

	apiClient := api.NewClientFromHTTP(api.AuthHTTPClient(hostname))
	org, _ := api.CurrentUserOrganization(apiClient, hostname)

	// TODO: define version/kind from manifest objective?
	objectiveResults, err := api.GetObjectiveResults(apiClient, entityHost, apiVersion, org.Name)
	if err != nil {
		return fmt.Errorf("failed to get objective results: %w", err)
	}
	_ = objectiveResults

	filteredObjectiveResults := filterObjectivesResults(objectiveResults, objectives, reportsLimit)

	// TODO: !!Important!! at the moment each objective result represents it's status when
	// the indicator was pushed vs. the objective at that time. If the objective is updated, only
	// indicators after it will produce an objective result with the delta of the new one. An alternative is to always
	// Use the latest objective against objective results. So, either an objective change updates previous objective results
	// retrospectively, or each stands on its own.
	reports, err := mapToReports(filteredObjectiveResults, reportsLimit, apiVersion)
	if err != nil {
		return fmt.Errorf("failed to generate report: %w", err)
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
		report.Write(out.Format, reports[0], w, log.StandardLogger(), reports[1], editReportSlice(reports[1:]))

		if outfile, ok := w.(*os.File); ok {
			outfile.Close() // explicitly closing the file handle
		}

	}

	return nil
}

func editReportSlice(s []*report.Report) *[]report.Report {
	var sNew []report.Report
	for _, v := range s {
		sNew = append(sNew, *v)
	}
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
	_ = filteredObjResClean
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

func mapToReports(objResults [][]entities.ObjectiveResultResponse, limit int, apiVersion string) ([]*report.Report, error) {
	var mappedReports []*report.Report
	_ = mappedReports

	for i := 0; i <= limit; i++ {
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
			if _, ok := serviceList[serviceLabel][nameLabel]; !ok {
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
					layout := "2006-01-02 15:04:05.000 +0000 UTC"
					to, err := time.Parse(layout, sl[i].Metadata.Labels["to"])
					if err != nil {
						return nil, fmt.Errorf("time 'to' not parsed correctly: %w", err)
					}
					from, err := time.Parse(layout, sl[i].Metadata.Labels["from"])
					if err != nil {
						return nil, fmt.Errorf("time 'from' not parsed correctly: %w", err)
					}
					// TODO: remove this once entity server returns period or calculated by from/to
					period, _ := iso8601.FromString("P0Y0DT1H0M0S")

					service.ServiceLevels = append(service.ServiceLevels, &report.ServiceLevel{
						Name:      name,
						Type:      sl[i].Spec.IndicatorSelector["category"],
						Objective: sl[i].Spec.ObjectivePercent,
						// TODO: get period from Entity Server, or is this calculated by to-from?!
						Period: core.Iso8601Duration{
							Duration: *period,
						},
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

			}

			mappedReports[i].Services = append(mappedReports[i].Services, &service)
		}

	}

	return mappedReports, nil
}
