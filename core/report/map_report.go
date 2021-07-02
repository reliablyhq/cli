package report

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"sort"
	"strings"
	"time"

	iso8601 "github.com/ChannelMeter/iso8601duration"
	"github.com/reliablyhq/cli/api"
	"github.com/reliablyhq/cli/config"
	"github.com/reliablyhq/cli/core"
	"github.com/reliablyhq/cli/core/entities"
	"github.com/reliablyhq/cli/core/iostreams"
	"github.com/reliablyhq/cli/core/manifest"
	"github.com/reliablyhq/cli/utils"
)

type ReportOutput struct {
	Format Format
	Path   string
}

type ReportOptions struct {
	IO *iostreams.IOStreams

	Selector      string
	ManifestPath  string
	OutputPath    string
	OutputFormat  string
	TemplateFile  string
	WatchFlag     bool
	OutputPaths   []string
	OutputFormats []string
	Service       string

	Outputs []ReportOutput
}

func EditReportSlice(s []*Report) *[]Report {
	var sNew []Report
	for _, v := range s {
		sNew = append(sNew, *v)
	}
	utils.Reverse(sNew)
	return &sNew
}

// func MapToReports(objResults [][]entities.ObjectiveResultResponse, limit int, apiVersion string) ([]*Report, error) {
func MapToReports(objectives []api.ExpandedObjective, limit int, apiVersion string) ([]*Report, error) {
	var mappedReports []*Report

	for i := 0; i < limit; i++ {

		var services []*Service = make([]*Service, 0)
		mappedReports = append(mappedReports, &Report{
			APIVersion: apiVersion,
			Timestamp:  time.Now().UTC(),
			Services:   services,
		})

		// Map into services. If no service, use "-" for now
		serviceList := make(map[string][]api.ExpandedObjective)
		for _, obj := range objectives {
			serviceName, ok := obj.Metadata.Labels["service"]
			if !ok {
				serviceName = " "
			}
			if _, ok := serviceList[serviceName]; !ok {
				serviceList[serviceName] = make([]api.ExpandedObjective, 0)
			}
			serviceList[serviceName] = append(serviceList[serviceName], obj)
		}

		for serviceLabel, s := range serviceList {

			serviceLevels := make([]*ServiceLevel, 0)
			service := Service{
				Name:          serviceLabel,
				Dependencies:  []string{},
				ServiceLevels: serviceLevels,
			}

			for _, obj := range s {
				name, ok := obj.Metadata.Labels["name"]
				if !ok {
					continue
				}
				objRes := obj.ForEach.ObjectiveResults
				if len(objRes) > i {
					sloIsMet := false
					if objRes[i].Spec.RemainingPercent >= 0 {
						sloIsMet = true
					}
					to, err := isoTimeParse(objRes[i].Metadata.Labels["to"])
					if err != nil {
						return nil, fmt.Errorf("time 'to' not parsed correctly: %w", err)
					}
					from, err := isoTimeParse(objRes[i].Metadata.Labels["from"])
					if err != nil {
						return nil, fmt.Errorf("time 'from' not parsed correctly: %w", err)
					}
					// Remove this once entity server returns period or calculated by from/to
					// timeDiff := to.Sub(from)
					// period := toIso8601Duration(timeDiff)
					period, err := time.ParseDuration(obj.Spec.Window.String())
					if err != nil {
						return nil, fmt.Errorf("time window cannot be parsed: %w", err)
					}
					isoPeriod := toIso8601Duration(period)

					service.ServiceLevels = append(service.ServiceLevels, &ServiceLevel{
						Name:      name,
						Type:      objRes[i].Spec.IndicatorSelector["category"],
						Objective: objRes[i].Spec.ObjectivePercent,
						Period:    isoPeriod,
						Result: &ServiceLevelResult{
							Actual:   objRes[i].Spec.ActualPercent,
							Delta:    objRes[i].Spec.RemainingPercent,
							SloIsMet: sloIsMet,
						},
						ObservationWindow: Window{
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

func GetReports(opts *ReportOptions) ([]*Report, error) {

	reportsLimit := 5
	apiVersion := "reliably.com/v1"

	var manObjectives entities.Manifest
	// Temporarily detecting old manifest
	if opts.ManifestPath != "" {
		isOld := IsDeprecatedManifest(opts.ManifestPath)
		if isOld {
			return nil, fmt.Errorf(
				"manifest '%s' is using a deprecated format. Please generate a new one with `reliably slo init`",
				opts.ManifestPath,
			)
		}

		// Will use this to filter results based on manifest objectives
		err := manObjectives.LoadFromFile(opts.ManifestPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read manifest: %w", err)
		}
		if len(manObjectives) < 1 {
			return nil, fmt.Errorf("no objectives found")
		}
	}

	// Parse selectors
	var selector map[string]string
	if opts.Selector != "" {
		var err error
		selector, err = ParseSelectors(opts.Selector)
		if err != nil {
			return nil, err
		}
	}

	hostname := config.Hostname
	entityHost := config.EntityServerHost
	apiClient := api.NewClientFromHTTP(api.AuthHTTPClient(hostname))
	org, err := config.GetCurrentOrgInfo()
	if err != nil {
		return nil, err
	}

	queryBody := api.QueryBody{
		Kind:   "objective",
		Labels: selector,
		Limit:  50,
		ForEach: api.ForEach{
			ObjectiveResult: api.ObjectiveResult{
				Include: true,
				Limit:   reportsLimit,
			},
		},
	}

	response, err := api.Query(apiClient, entityHost, apiVersion, org.Name, queryBody)
	if err != nil {
		return nil, fmt.Errorf("failed to get objective results: %w", err)
	}

	// TODO: if manifest objectives, filter list of objectives to have only ones that match that one.
	// Can use the entity server id function to generate id from one and then the other and compare and filter maybe?
	reports, err := MapToReports(response.Objectives, reportsLimit, apiVersion)
	if err != nil {
		return nil, fmt.Errorf("failed to generate report: %w", err)
	}

	return reports, nil
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

func ClearScreen() {
	var c *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		c = exec.Command("cmd", "/c", "cls")
	default:
		// clear should work for UNIX & linux based systems
		c = exec.Command("clear")

		// hide cursor on unix based systems
		fmt.Print("\033[?25l")
	}

	c.Stdout = os.Stdout
	c.Run()
}

func IsDeprecatedManifest(path string) bool {
	m, _ := manifest.Load(path)
	if m != nil {
		return m.Services != nil
	}
	return false
}

func ParseSelectors(s string) (map[string]string, error) {
	selector := make(map[string]string)
	commaSep := strings.Split(s, ",")
	for _, v := range commaSep {
		equalSep := strings.Split(v, "=")
		if len(equalSep) != 2 {
			return nil, fmt.Errorf("selector string incorrectly delimited")
		}
		selector[equalSep[0]] = equalSep[1]
	}
	return selector, nil
}

// Function filters objectiveResultResponses based on objectives.
// It is assumed the slice is ordered in descending order by creation time.
// If Entity Server assumes filtering, this function can likely be replaced.
// func filterObjectivesResults(
// 	objectiveResults *[]entities.ObjectiveResultResponse,
// 	objectives []*entities.Objective,
// 	maxResults int,
// ) [][]entities.ObjectiveResultResponse {

// 	// Creating a hash table for faster search, although likely not many entities.
// 	// Array of name+service used as map key.
// 	objectivesMapped := make(map[[2]string]int)
// 	for i, o := range objectives {
// 		if _, ok := o.Metadata.Labels["name"]; !ok {
// 			continue
// 		}
// 		if _, ok := o.Metadata.Labels["service"]; !ok {
// 			continue
// 		}
// 		objectivesMapped[[2]string{
// 			o.Metadata.Labels["name"], o.Metadata.Labels["service"],
// 		}] = i
// 	}

// 	filteredObjRes := make([][]entities.ObjectiveResultResponse, len(objectives))
// 	for _, or := range *objectiveResults {
// 		if _, ok := or.Metadata.Labels["name"]; !ok {
// 			continue
// 		}
// 		if _, ok := or.Metadata.Labels["service"]; !ok {
// 			continue
// 		}
// 		nameAndService := [2]string{or.Metadata.Labels["name"], or.Metadata.Labels["service"]}
// 		if objectiveIndex, ok := objectivesMapped[nameAndService]; ok {
// 			if len(filteredObjRes[objectiveIndex]) <= maxResults {
// 				filteredObjRes[objectiveIndex] = append(filteredObjRes[objectiveIndex], or)
// 			}
// 		}
// 	}

// 	filteredObjResClean := make([][]entities.ObjectiveResultResponse, 0)
// 	for _, v := range filteredObjRes {
// 		if v != nil {
// 			filteredObjResClean = append(filteredObjResClean, v)
// 		}
// 	}

// 	return filteredObjResClean
// }
