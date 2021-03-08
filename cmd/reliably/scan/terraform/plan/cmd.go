package plan

import (
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/reliablyhq/cli/core"
	"github.com/reliablyhq/cli/core/terraform"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const platform = "terraform"

var (
	file string
)

// New returns a new plan command
func New() *cobra.Command {
	cmd := cobra.Command{
		Use:   "plan",
		Short: "scan a terraform plan",
		Long:  "scan a terraform plan for policy violations",
		Run:   run,
	}

	cmd.Flags().StringVarP(&file, "file", "f", "", "the path to the terraform plan")

	return &cmd
}

func run(cmd *cobra.Command, args []string) {
	if file == "" {
		log.Error("file argument is required")
		os.Exit(1)
	}

	// 1: check for the file
	contentBytes, err := getFileContent(file)
	if err != nil {
		log.Debug(err)
		log.Error("An error occured while opening the file")
		os.Exit(1)
	}

	// 2: parse the file content
	var tfPlan terraform.PlanRepresentation
	if err := json.Unmarshal(contentBytes, &tfPlan); err != nil {
		log.Debug(err)
		log.Error("An error occured while trying to deserailize the content of the file")
		os.Exit(1)
	}

	// 3: download policies
	policyMap := map[string]string{}
	var downloadFunc func(*terraform.ModuleRepresentation)
	downloadFunc = func(module *terraform.ModuleRepresentation) {
		for _, resource := range module.Resources {
			kind := resource.Type
			path, err := core.FetchPolicy(".reliably", platform, kind)
			if err != nil {
				log.Warn(err)
				continue
			}

			policyMap[kind] = path
		}

		for _, childModule := range module.ChildModules {
			downloadFunc(childModule.ToModule())
		}
	}

	downloadFunc(tfPlan.PlannedValues.RootModule)

	// evaluate the resources
	var evalFunc func(*terraform.ModuleRepresentation)
	evalFunc = func(module *terraform.ModuleRepresentation) {
		for _, res := range module.Resources {
			kind := res.Type
			if policyPath, ok := policyMap[kind]; ok {
				resultSet := core.Eval(policyPath, res.Values)
				core.PrintViolations(resultSet, policyPath, platform, kind, 0)
			}
		}

		for _, childModule := range module.ChildModules {
			evalFunc(childModule.ToModule())
		}
	}

	evalFunc(tfPlan.PlannedValues.RootModule)

	// // 4: download policies
	// policyResources := make(map[string]*core.Resource)

	// for _, res := range resources {
	// 	pol, err := core.FetchPolicy(".reliably", platform, res.Kind)
	// 	if err != nil {
	// 		log.Warn(err)
	// 		continue
	// 	}

	// 	if pol != "" {
	// 		policyResources[pol] = res
	// 	}
	// }

	// // 5: analyse
	// if len(policyResources) == 0 {
	// 	log.Info("no policies found for the scanned resources")
	// 	os.Exit(1)
	// }

	// for pol, res := range policyResources {
	// 	rs := core.Eval(pol, res)
	// 	core.PrintViolations(rs, pol, platform, res.Kind, res.StartingLine)
	// }
}

func getFileContent(path string) ([]byte, error) {
	return ioutil.ReadFile(path)
}
