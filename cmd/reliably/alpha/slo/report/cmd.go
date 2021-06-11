package reportAlpha

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/reliablyhq/cli/api"
	sloReport "github.com/reliablyhq/cli/cmd/reliably/slo/report"
	"github.com/reliablyhq/cli/core"
	"github.com/reliablyhq/cli/core/entities"
	"github.com/reliablyhq/cli/core/report"
	v "github.com/reliablyhq/cli/version"
	log "github.com/sirupsen/logrus"
)

func AlpaReportRun(opts *sloReport.ReportOptions) error {
	// TODO: implement watch for entity server
	// check for -w/--watch
	// if opts.WatchFlag {
	// 	return watch(opts)
	// }

	var sliSelect entities.Labels = entities.Labels{
		"provider":       "gcp",
		"category":       "availability",
		"gcp_project_id": "abc123",
		"resource_id":    "projectid/google-cloud-load-balancers/loadbalancer-name",
	}

	slo1 := entities.Objective{
		TypeMeta: entities.TypeMeta{APIVersion: "v1", Kind: "Objective"},
		Metadata: entities.Metadata{
			Labels: map[string]string{
				"name":    "api-availability",
				"service": "example-api",
			},
		},
		Spec: entities.ObjectiveSpec{
			IndicatorSelector: entities.Selector(sliSelect),
			ObjectivePercent:  90,
			Window:            core.Duration{Duration: time.Duration(24 * time.Hour)},
		},
	}

	slo2 := entities.Objective{
		TypeMeta: entities.TypeMeta{APIVersion: "v1", Kind: "Objective"},
		Metadata: entities.Metadata{
			Labels: map[string]string{
				"name":    "api-latency",
				"service": "example-api",
			},
		},
		Spec: entities.ObjectiveSpec{
			IndicatorSelector: entities.Selector(sliSelect),
			ObjectivePercent:  90,
			Window:            core.Duration{Duration: time.Duration(24 * time.Hour)},
		},
	}

	slos := []entities.Objective{slo1, slo2}

	// TODO: swap to the method that reads new manifest, instead of hard coded above
	// m, err := manifest.Load(manifestPath)
	// if err != nil {
	// 	return fmt.Errorf("failed to read manifest: %w", err)
	// }
	// // END

	opts.IO.StartProgressIndicator()

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

	// TODO: replace with the previous objectiveResults for each objective
	var lr report.Report
	var reports []report.Report

	// TODO: process objective results, filter by name and service in labels.
	// Validate for when it doesn't have these two labels.
	// TODO: map into report.Report type!
	_ = objectiveResults
	// 	// ObjectiveResults are ordered by created at on the backend.
	// 	// So you can iterate over the response, for each object add to another slice that has the results for that object.
	// 	// Essentially a slice of slices of objectiveresults. you then take the first objectiveresult for each slice, the rest for trends.
	// 	// map this to the current report structure. (see below though)
	// 	for i := range slos {
	// 		_ = i
	// 	}
	// TODO: map into report.Report type!

	lr = reports[0]

	// !! TODO: need to basically reimplement this below so it's from objectiveresult -> report type.
	// Currently it does the querying of the providers, etc... so this part you don't need to consider
	// Could be called something like "FromObjectiveResults" !!
	r, err := report.FromManifest(m)
	if err != nil {
		return err
	}

	// TODO: likely remove this bit, not sure yet
	// // reverse last reports - oldest to most recent - append current at the end
	// utils.Reverse(reports)
	// reports = append(reports, *r)

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
		report.Write(out.Format, r, w, log.StandardLogger(), &lr, &reports)

		if outfile, ok := w.(*os.File); ok {
			outfile.Close() // explicitly closing the file handle
		}

	}

	return nil
}
