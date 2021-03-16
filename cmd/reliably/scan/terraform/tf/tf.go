package tf

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/hashicorp/hcl"
	"github.com/reliablyhq/cli/core"
	"github.com/reliablyhq/cli/scan"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const platform = "terraform"

// Resource -type used to create resource input for scanning
type Resource struct {
	Label  string                 `json:"label"`
	Type   string                 `json:"type"`
	Values map[string]interface{} `json:"values"`
}

// New returns a new plan command
func New() *cobra.Command {
	cmd := cobra.Command{
		Use: "tf",
		Example: strings.Join([]string{
			"reliably scan terraform tf .",
			"reliably scan terraform tf ./path/to/terraform/dir",
			"reliably scan terraform tf ./path/to/resources.tf",
		}, "\n"),
		Short: "scan a terraform .tf manifest(s)",
		Run:   run,
	}
	return &cmd
}

func run(cmd *cobra.Command, args []string) {
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
			log.Error("An error occured while looking up tfiles in directory")
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
			log.Error("An error occured while opening the file")
			os.Exit(1)
		}

		var resources map[string]interface{}
		if err := unmarshalTF(b, &resources); err != nil {
			log.Debug(err)
			log.Error("An error occured while parsing the file")
			os.Exit(1)
		}

		// 3. structure resource and scan
		for _, resourceMap := range resources["resource"].([]map[string]interface{}) {
			for resourceType, v := range resourceMap {
				// resource := make(map[string]interface{})
				var resource Resource
				resource.Type = resourceType
				for label, r := range v.([]map[string]interface{})[0] {
					resource.Label = label
					resource.Values = r.([]map[string]interface{})[0]
				}
				target := scan.NewTarget(&resource, platform, resourceType)
				result, err := scan.FindPolicyAndEvaluate(target)
				if err != nil {
					log.Debug(err)
					log.Warnf("An error occured while scanning resource: [%s]", resource.Type)
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

	log.Info("Scan complete!")
}

// unmarshal .tf files by recursively removing
// unsupported lines. The theory being that
// we'll only be interested values we can directly support
func unmarshalTF(tf []byte, v interface{}) error {
	var re = regexp.MustCompile(`Unknown token\:.(\d{1,5}):`)

	if err := hcl.Unmarshal(tf, v); err != nil {
		if !re.MatchString(err.Error()) {
			return err
		}

		ln := re.FindAllStringSubmatch(err.Error(), -1)[0][1]
		badlineNo, _ := strconv.Atoi(ln)

		var updated strings.Builder
		var lineNo = 1
		scanner := bufio.NewScanner(strings.NewReader(string(tf)))

		// remove lineNo
		for scanner.Scan() {
			if lineNo != badlineNo {
				fmt.Fprintln(&updated, scanner.Text())
			}
			lineNo++
		}
		unmarshalTF([]byte(updated.String()), v)
	}
	return nil
}
