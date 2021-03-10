package plan

import (
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/icza/dyno"
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
		Run:   run,
	}

	cmd.Flags().StringVarP(&file, "file", "f", "", "the path to the terraform plan")

	return &cmd
}

func run(cmd *cobra.Command, args []string) {
	log.Info("This hasn't been implemented yet. Check back in a later version to see if its ready!")
	os.Exit(1)
}

func runWip(cmd *cobra.Command, args []string) {
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
	var evalFunc func(*terraform.ModuleRepresentation)
	evalFunc = func(module *terraform.ModuleRepresentation) {
		for _, resource := range module.Resources {
			kind := resource.Type
			log.Infof("Analysing %s...", kind)

			path, err := core.FetchPolicy(".reliably", platform, kind)
			if err != nil {
				log.Warn(err)
				continue
			}

			if path == "" {
				log.Warnf("policy not found for resource '%s'", kind)
				continue
			}

			log.Debugf("Policy found for %s. Processing rules...", kind)

			input := dyno.ConvertMapI2MapS(resource)
			resultSet := core.Eval(path, input)
			violationCount := core.CountViolations(resultSet, platform, kind)
			log.Infof("Found %v violations", violationCount)

			if violationCount > 0 {
				core.PrintViolations(resultSet, path, platform, kind, 0)
			}

			log.Infof("Processing %s complete!", kind)
		}

		for _, childModule := range module.ChildModules {
			log.WithField("module", childModule.Address)
			evalFunc(childModule.ToModule())
		}
	}

	log.WithField("module", "root")
	evalFunc(tfPlan.PlannedValues.RootModule)

	log.Info("Scan complete!")
}

func getFileContent(path string) ([]byte, error) {
	return ioutil.ReadFile(path)
}
