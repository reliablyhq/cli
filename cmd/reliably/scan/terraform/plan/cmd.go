package plan

import (
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/reliablyhq/cli/core/terraform"
	"github.com/reliablyhq/cli/scan"
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
	// if file == "" {
	// 	log.Error("file argument is required")
	// 	os.Exit(1)
	// }

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

			target := scan.NewTarget(resource, platform, resource.Type)
			result, err := scan.FindPolicyAndEvaluate(target)
			if err != nil {
				log.Debug(err)
				log.Warnf("error occured while scanning target: %s", target.ResourceType)
				continue
			}

			for _, r := range result.Violations {
				// log.Infof("Resource: [%s] - Message: %s", target.ResourceType, r.Message)
				log.Infof("[%s] [%s]: %s", target.ResourceType, resource.Name, r.Message)
			}

			log.Infof("Processing %s complete!", target.ResourceType)
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
