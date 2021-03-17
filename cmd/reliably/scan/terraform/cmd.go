package terraform

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/reliablyhq/cli/core"
	"github.com/reliablyhq/cli/core/terraform"
	"github.com/reliablyhq/cli/scan"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var plan string

// New terraform command
func New() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "terraform",
		Short: "scan terraform resources",
		Long:  "scan terraform resources for policy violations",
		Example: strings.Join([]string{
			"reliably scan terraform .",
			"reliably scan terraform ./path/to/terraform/dir",
			"reliably scan terraform ./path/to/resources.tf",
			"reliably scan terraform --plan path/to/plan.json",
		}, "\n"),
		Run: run,
	}

	cmd.Flags().StringVarP(&plan, "plan", "p", "", "the path to the terraform plan")
	return cmd
}

func run(cmd *cobra.Command, args []string) {

	// check if --plan flag is used. and call scanPlan
	// else scan for TF files
	if plan != "" {
		log.Debugf("scanning terraform plan: %s", plan)
		scanPlan()
	} else {
		scanTF(args)
	}

	log.Info("Scan complete!")
}

// scanTF - scans terraform files (.tf)
func scanTF(args []string) {
	var (
		tfiles []string
		err    error
	)

	// 1. determine tf file(s) with direct args, A single file or dir can be passed
	var dir = "."
	if len(args) != 0 {
		dir = args[0]
	}

	if strings.HasSuffix(strings.ToLower(dir), ".tf") {
		tfiles = []string{dir}
	} else {
		log.Debugf("checking for (.tf) files in: [%s]", filepath.Join(dir, "*.tf"))
		tfiles, err = filepath.Glob(filepath.Join(dir, "*.tf"))
		if err != nil {
			log.Debug(err)
			log.Error("An error occurred while looking up tfiles in directory")
			os.Exit(1)
		}
	}

	if len(tfiles) == 0 {
		log.Info("no terraform (.tf) files detected")
		os.Exit(0)
	}

	// 2. Read and Unmarshal files in sequence
	for _, f := range tfiles {
		b, err := ioutil.ReadFile(f)
		if err != nil {
			log.Debug(err)
			log.Error("An error occurred while opening the file")
			os.Exit(1)
		}

		var resources map[string]interface{}
		if err := terraform.UnmarshalTF(b, &resources); err != nil {
			log.Debug(err)
			log.Error("An error occurred while parsing the file")
			os.Exit(1)
		}

		// 3. structure resource and scan
		for _, resourceMap := range resources["resource"].([]map[string]interface{}) {
			for resourceType, v := range resourceMap {
				// resource := make(map[string]interface{})
				var resource terraform.TFResource
				resource.Type = resourceType
				for label, r := range v.([]map[string]interface{})[0] {
					resource.Label = label
					resource.Values = r.([]map[string]interface{})[0]
				}
				target := scan.NewTarget(&resource, terraform.Platform, resourceType)
				result, err := scan.FindPolicyAndEvaluate(target)
				if err != nil {
					log.Debug(err)
					log.Warnf("An error occurred while scanning resource: [%s]", resource.Type)
					continue
				}

				// 4. print result for a given resource, (if any)
				for _, r := range result.Violations {
					log.Infof("[%s] [%s] [%s] [%s]: %s", core.Level(r.Level), f, target.ResourceType, resource.Label, r.Message)
				}

				log.Infof("Processing %s complete!", target.ResourceType)
			}
		}

	}
}

// scanPlan - scan terraform JSON plan files.
// this function uses the -p/--plan flag explicitly
func scanPlan() {
	// 1: check for the file
	b, err := ioutil.ReadFile(plan)
	if err != nil {
		log.Debug(err)
		log.Error("An error occurred while opening the file")
		os.Exit(1)
	}

	// 2: parse the file content
	var tfPlan terraform.PlanRepresentation
	if err := json.Unmarshal(b, &tfPlan); err != nil {
		log.Debug(err)
		log.Error("An error occurred while trying to deserailize the content of the file")
		os.Exit(1)
	}

	// 3: download policies
	var evalFunc func(*terraform.ModuleRepresentation)
	evalFunc = func(module *terraform.ModuleRepresentation) {
		for _, resource := range module.Resources {

			target := scan.NewTarget(resource, terraform.Platform, resource.Type)
			result, err := scan.FindPolicyAndEvaluate(target)
			if err != nil {
				log.Debug(err)
				log.Warnf("error occurred while scanning target: %s", target.ResourceType)
				continue
			}

			for _, r := range result.Violations {
				// log.Infof("Resource: [%s] - Message: %s", target.ResourceType, r.Message)
				log.Infof("[%s] [%s] [%s]: %s", core.Level(r.Level), target.ResourceType, resource.Name, r.Message)
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
}
