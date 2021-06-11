package reportAlpha

import (
	"errors"
	"fmt"
	"os"

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
	objectiveResults, err := api.GetObjectiveResults(apiClient, entityHost, "v1", "objective-results", org.Name)
	if err != nil {
		return fmt.Errorf("failed to get objective results: %w", err)
	}
	_ = objectiveResults

	filteredObjectiveResults := filterObjectivesResults(objectiveResults, objectives, 6)
	_ = filteredObjectiveResults
	// 2. Translate this slice of slices into a report type similar to FromManifest

	var reports []report.Report
	// 3. Complete below

	// ..
	// ..
	// ..
	// ..
	// TODO: likely remove this bit, not sure yet
	// // reverse last reports - oldest to most recent - append current at the end
	// utils.Reverse(reports)
	// reports = append(reports, *r)
	lr := reports[0]
	_ = lr
	opts.IO.StopProgressIndicator()

	// for _, out := range opts.Outputs {

	// 	var w io.Writer = os.Stdout
	// 	if out.Path != "" {
	// 		outfile, err := os.Create(out.Path) // creates or truncates with O_RDWR mode
	// 		if err != nil {
	// 			log.Error("error creating output file")
	// 			log.Error(err)
	// 			return err
	// 		}
	// 		w = outfile
	// 		// we cannot defer outfile closing here as we are in a for-loop
	// 	}
	// 	report.Write(out.Format, r, w, log.StandardLogger(), &lr, &reports)

	// 	if outfile, ok := w.(*os.File); ok {
	// 		outfile.Close() // explicitly closing the file handle
	// 	}

	// }

	return nil
}

// Function filters objectiveResultResponses based on objectives.
// It is assumed the slice is ordered in descending order by creation time.
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
