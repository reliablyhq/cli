/*
package kubernetes is a colletion of funtions for interactng with a live
kubernetes cluster

*/
package kubernetes

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/serializer/json"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type KubernetesAPI struct {
	APIVersion string `yaml:"apiversion"`
	Kind       string `yaml:"kind"`
	Metadata   struct {
		Name      string `yaml:"name"`
		Namespace string `yaml:"namespace"`
		// Labels    struct {
		// 	Source string `yaml:"source"`
		// } `yaml:"labels"`
	} `yaml:"metadata"`
}

func (m KubernetesAPI) URI() string {
	var uri string

	if m.Metadata.Namespace != "" {
		uri = fmt.Sprintf("/api/%s/%s/%s/%s", m.APIVersion, m.Kind, m.Metadata.Namespace, m.Metadata.Name)
	} else {
		uri = fmt.Sprintf("/api/%s/%s/%s", m.APIVersion, m.Kind, m.Metadata.Name)
	}

	return strings.ToLower(uri)
}

func GetYamlInfo(yamlContent string) (*KubernetesAPI, error) {

	var m KubernetesAPI
	err := yaml.Unmarshal([]byte(yamlContent), &m)
	if err != nil {
		return nil, fmt.Errorf("Could not unmarshal: %v \n---\n%v", err, yamlContent)
	}
	// fmt.Printf("unMarshalled content %v", m)

	// if m.Kind == "" {
	// 	fmt.Println("yaml file with kind missing")

	// 	return nil, fmt.Errorf("yaml file with kind missing")
	if m.Metadata.Name == "" {
		fmt.Println("yaml file with name missing")
		return nil, fmt.Errorf("yaml file with name missing")
	}

	return &m, nil
}

// ConnectToKubernetes uses the default kubectl config file to create a
// Clientset for the default config
func ConnectToKubernetes() (*kubernetes.Clientset, error) {
	// Pull the config
	config, err := clientcmd.BuildConfigFromFlags("", os.Getenv("HOME")+"/.kube/config")
	// Connect
	cs, err := kubernetes.NewForConfig(config)
	return cs, err
}

// GetPods provide a list an of JSON Pod specs from the clientset
func GetPods(cs kubernetes.Clientset, namespace string) (po []string, err error) {
	pods, err := cs.CoreV1().Pods(namespace).List(metav1.ListOptions{})
	if err != nil {
		return
	}
	for _, p := range pods.Items {
		var podYAML strings.Builder

		// writing the "Kind" property as this is not included
		// in the serialized yaml output but may be required
		// for discovery
		fmt.Fprintln(&podYAML, "kind: Pod")

		e := json.NewYAMLSerializer(json.DefaultMetaFactory, nil, nil)
		e.Encode(p.DeepCopyObject(), &podYAML)
		po = append(po, podYAML.String())
	}
	return po, err
}
