package tf

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/hashicorp/hcl"
	"github.com/reliablyhq/cli/scan"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const platform = "terraform"

var file string

// Resource -type used to create resource input for scanning
type Resource struct {
	Type   string                 `json:"type"`
	Values map[string]interface{} `json:"values"`
}

// New returns a new plan command
func New() *cobra.Command {
	cmd := cobra.Command{
		Use:   "tf",
		Short: "scan a terraform .tf manifest",
		Run:   run,
	}

	cmd.Flags().StringVarP(&file, "file", "f", "", "the path to the terraform (.tf) file")
	return &cmd
}

func run(cmd *cobra.Command, args []string) {

	b, err := ioutil.ReadFile(file)
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

	for _, resourceMap := range resources["resource"].([]map[string]interface{}) {
		for resourceType, v := range resourceMap {
			// resource := make(map[string]interface{})
			var resource Resource
			resource.Type = resourceType
			for _, r := range v.([]map[string]interface{})[0] {
				resource.Values = r.([]map[string]interface{})[0]
			}
			target := scan.NewTarget(&resource, platform, resourceType)
			result, err := scan.FindPolicyAndEvaluate(target)
			if err != nil {
				log.Debug(err)
				log.Errorf("An error occured while scanning resource: [%s]", resource.Type)
				continue
			}

			for _, r := range result.Violations {
				log.Infof("[%s]: %s", target.ResourceType, r.Message)
			}

			log.Infof("Processing %s complete!", target.ResourceType)
		}
	}

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
