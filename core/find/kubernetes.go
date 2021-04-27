package find

import (
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"

	"github.com/karrick/godirwalk"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

const (
	kindRegex = `(?m)^[kK]ind:\s?(.*)$`
)

// KubernetesAPI is a minimal struct for unmarshaling kubernetes configs into
type KubernetesAPI struct {
	APIVersion string `yaml:"apiVersion"`
	Kind       string `yaml:"kind"`
	Metadata   struct {
		Name      string `yaml:"name"`
		Namespace string `yaml:"namespace"`
		Labels    struct {
			Source string `yaml:"source"`
		} `yaml:"labels"`
	} `yaml:"metadata"`
}

// URI returns a unique identifier referencing the resource
// It uses the following patterns:
// /api/<apiversion>/<kind>/<name>
// /api/<apiversion>/<kind>/<namespace>/<name>
func (m KubernetesAPI) URI() string {
	var uri string

	if m.Metadata.Namespace != "" {
		uri = fmt.Sprintf("/api/%s/%s/%s/%s", m.APIVersion, m.Kind, m.Metadata.Namespace, m.Metadata.Name)
	} else {
		uri = fmt.Sprintf("/api/%s/%s/%s", m.APIVersion, m.Kind, m.Metadata.Name)
	}

	return strings.ToLower(uri)
}

// GetKubernetesFiles returns the list of Kubernetes files found
// on the local file system recursively starting at a base directory
func GetKubernetesFiles(baseDirectory string) []string {
	log.Debug("Scanning directory '", baseDirectory, "' for Kubernetes files")
	var files = make([]string, 0)
	_ = godirwalk.Walk(baseDirectory, &godirwalk.Options{
		Callback: func(osPathname string, de *godirwalk.Dirent) error {
			if (osPathname != "." && strings.HasPrefix(osPathname, ".") && !strings.HasPrefix(osPathname, "..")) || // ignore hidden files/folders but not parent folder ..
				strings.Contains(osPathname, ".git") ||
				strings.Contains(osPathname, ".github") ||
				strings.Contains(osPathname, ".circleci") ||
				strings.Contains(osPathname, ".travis") ||
				strings.Contains(osPathname, ".reliably") {
				//log.Debug(fmt.Sprintf("Skip %v", osPathname))
				return godirwalk.SkipThis
			}

			if strings.HasSuffix(osPathname, ".yaml") ||
				strings.HasSuffix(osPathname, ".yml") {
				// open the yaml file to detect its kind

				log.WithFields(log.Fields{
					"path": osPathname,
				}).Debug("Yaml file found")
				files = append(files, osPathname)
			}
			return nil
		},
		Unsorted: true,
	})
	return files
}

// ExtractKindFromManifest tries to extract the Kind value from a
// Kubernetes resource string
// It returns an empty string, if the Kind cannot be found, or if the
// string value is not a valid Kubernetes resource
func ExtractKindFromManifest(manifest string) string {
	re := regexp.MustCompile(kindRegex)
	matches := re.FindStringSubmatch(manifest)

	if len(matches) > 0 {
		return matches[1]
	}

	return ""
}

// Resource ...
type Resource struct {
	kind        string
	content     string
	startAtLine int
}

// Manifest ...
type Manifest struct {
	filename  string
	resources []Resource
}

/*
func SplitManifestPerResource() []Resource {

	var resources []Resource

	m := Manifest{filename: "ertyu"}

	return resources
}
*/

// ReadAndSplitKubernetesFile reads and split a single manifest
// file that may contain several Kubernetes resources into an array of
// resources (as string)
// If the file contains a single K8s resource, the returned array
// will contain a single element
func ReadAndSplitKubernetesFile(file string) []string {
	var fileContent []byte

	if file == "-" {
		defer os.Stdin.Close()
		c, err := io.ReadAll(os.Stdin)
		if err != nil {
			log.Fatalf("Failed reading from stdin : %v", err)
		}
		fileContent = c
	} else {
		c, err := os.ReadFile(file)
		if err != nil {
			log.Fatalf("Failed reading file %s : %v", file, err)
		}
		fileContent = c
	}

	docs := strings.Split(string(fileContent), "\n---")

	return docs

	/*
		res := []string{}
		// Trim whitespace in both ends of each yaml docs.
		// - Re-add a single newline last
		for _, doc := range docs {
			content := strings.TrimSpace(doc)
			// Ignore empty docs
			if content != "" {
				res = append(res, content+LineBreak)
			}
		}
		return res
	*/
}

// GetYamlInfo unmarshal the YAML into a Kubernetes meta structure
// that is usefull to ensure the unmarshaled YAML is a Kubernetes resource
// definition. The stucture will then contain basic K8S object meta
func GetYamlInfo(yamlContent string) (*KubernetesAPI, error) {

	var m KubernetesAPI
	err := yaml.Unmarshal([]byte(yamlContent), &m)
	if err != nil {
		return nil, fmt.Errorf("Could not unmarshal: %v \n---\n%v", err, yamlContent)
	}

	if m.Kind == "" {
		return nil, fmt.Errorf("yaml file with kind missing")
	} else if m.Metadata.Name == "" {
		return nil, fmt.Errorf("yaml file with name missing")
	}

	return &m, nil
}
