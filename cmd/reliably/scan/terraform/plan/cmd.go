package plan

import (
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/reliablyhq/cli/core/scanning"
	"github.com/reliablyhq/cli/types/terraform"
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
	resources := extractResources(&tfPlan)

	// 4: analyse
	var results []*scanning.Result
	for _, res := range resources {
		x, err := scanning.Scan(res)
		if err != nil {
			log.Warn(err)
			continue
		}

		results = append(results, x...)
	}

	// 4: print the outcome
	scanning.Print(results...)
}

func getFileContent(path string) ([]byte, error) {
	return ioutil.ReadFile(path)
}

// ExtractResources from the plan
func extractResources(p *terraform.PlanRepresentation) []*scanning.Resource {
	panic("not implemented")
}
