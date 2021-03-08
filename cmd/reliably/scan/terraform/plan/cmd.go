package plan

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"sync"

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

	// 3: extract resources
	resources, err := terraform.ExtractResourcesFromPlan(&tfPlan)
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}

	policyResources := make(map[string]*core.Resource)
	wg := sync.WaitGroup{}

	// 4: download policies
	wg.Add(len(resources))
	for _, res := range resources {
		go func(res *core.Resource) {
			defer wg.Done()
			pol, err := core.FetchPolicy(".reliably", platform, res.ID)
			if err == nil {
				log.Warn(err)
				return
			}

			if pol != "" {
				policyResources[pol] = res
			}
		}(res)
	}
	wg.Wait()

	// 5: analyse
	for pol, res := range policyResources {
		rs := core.Eval(pol, res)
		violations := core.ReportViolations(rs, pol, platform, res.ID, -1, "na", "na")
		log.Print(violations)
	}
}

func getFileContent(path string) ([]byte, error) {
	return ioutil.ReadFile(path)
}
