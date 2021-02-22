/*
package kubernetes is a colletion of funtions for interactng with a live
kubernetes cluster

*/
package kubernetes

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"

	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sJSON "k8s.io/apimachinery/pkg/runtime/serializer/json"
	"k8s.io/client-go/kubernetes"

	// auth plugin import allows auth mechanisms from kubernetes providers to function
	// i.e GCP, Azure, etc see: https://pkg.go.dev/k8s.io/client-go/plugin/pkg/client/auth
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/tools/clientcmd"
)

//pkg.go.dev/k8s.io/client-go/plugin/pkg/client/auth
// type KubernetesAPI struct {
// 	APIVersion string `yaml:"apiVersion"`
// 	Kind       string `yaml:"kind"`
// 	Metadata   struct {
// 		Name      string `yaml:"name"`
// 		Namespace string `yaml:"namespace"`
// 		Labels    struct {
// 			Source string `yaml:"source"`
// 		} `yaml:"labels"`
// 	} `yaml:"metadata"`
// }

type KubernetesAPI struct {
	APIVersion string `json:"apiVersion"`
	Kind       string `json:"kind"`
	Metadata   struct {
		Annotations struct {
		} `json:"annotations"`
		Labels struct {
			App string `json:"app"`
		} `json:"labels"`
		Name      string `json:"name"`
		Namespace string `json:"namespace"`
	} `json:"metadata"`
	Spec struct {
		Containers []struct {
			Image string `json:"image"`
			Name  string `json:"name"`
		} `json:"containers"`
		RestartPolicy string `json:"restartPolicy"`
	} `json:"spec"`
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

func GetHeaderInfo(content string) (*KubernetesAPI, error) {

	var m KubernetesAPI
	err := json.Unmarshal([]byte(content), &m)
	if err != nil {
		return nil, fmt.Errorf("Could not unmarshal: %v \n---\n%v", err, content)
	}

	if err != nil {
		return nil, fmt.Errorf("Could not unmarshal: %v \n---\n%v", err, content)
	}
	// fmt.Printf("unMarshalled content %v", m)

	if m.Kind == "" {
		fmt.Println("yaml file with kind missing")

		return nil, fmt.Errorf("yaml file with kind missing")
	}
	if m.Metadata.Name == "" {
		fmt.Println("yaml file with name missing")
		return nil, fmt.Errorf("yaml file with name missing")
	}

	return &m, nil
}

// GetKubernetesClientSet uses the default kubectl config file to create a
// Clientset for the default config
func GetKubernetesClientSet() (*kubernetes.Clientset, error) {
	// Pull the config
	// todo: deal with the case when you are not getting kube config from HOME location
	config, err := clientcmd.BuildConfigFromFlags("", os.Getenv("HOME")+"/.kube/config")
	// Connect
	clientSet, err := kubernetes.NewForConfig(config)
	return clientSet, err
}

// GetFormattedJSON takes a source string and outputs a formatted JSON
// string using 2 space characters as the indent
func GetFormattedJSON(source string) (result string) {
	var prettyJSON bytes.Buffer
	err := json.Indent(&prettyJSON, []byte(source), "", "  ")
	if err != nil {
		log.Debugf("Error formatting JSON: %v\n", source)
	}
	result = string(prettyJSON.Bytes())

	return result
}

// GetPodSpec provide a list an of JSON Pod specs from the clientset
func GetPodSpec(cs kubernetes.Clientset, namespace string) (po []string, err error) {
	pods, err := cs.CoreV1().Pods(namespace).List(metav1.ListOptions{})
	if err != nil {
		return
	}
	for _, p := range pods.Items {
		podJSON := regexp.MustCompile(`\n|\|`).
			ReplaceAllString(p.Annotations["kubectl.kubernetes.io/last-applied-configuration"], "")

		// if annotation was empty use k8s JSON serializer
		if len(podJSON) == 0 {
			var podRawJSON strings.Builder
			e := k8sJSON.NewSerializerWithOptions(k8sJSON.DefaultMetaFactory,
				nil, nil,
				k8sJSON.SerializerOptions{Yaml: false}, // Yaml: false, returns JSON
			)

			// Setting kind manually
			p.Kind = "Pod"

			// p.APIVersion = "v1"
			e.Encode(p.DeepCopyObject(), &podRawJSON)
			podJSON = podRawJSON.String()

			// log.Debugf("Error processing pod: %v\n", p.Name)

		}

		po = append(po, GetFormattedJSON(podJSON))

	}
	return po, err
}

// GetDeploymentSpec provide a list an of JSON Deployment specs from the clientset
func GetDeploymentSpec(cs kubernetes.Clientset, namespace string) (deploy []string, err error) {
	deployment, err := cs.AppsV1().Deployments(namespace).List(metav1.ListOptions{})
	if err != nil {
		return
	}
	for _, d := range deployment.Items {
		deployJSON := regexp.MustCompile(`\n|\|`).
			ReplaceAllString(d.Annotations["kubectl.kubernetes.io/last-applied-configuration"], "")

		if len(deployJSON) == 0 {
			var deployRawJSON strings.Builder
			e := k8sJSON.NewSerializerWithOptions(k8sJSON.DefaultMetaFactory,
				nil, nil,
				k8sJSON.SerializerOptions{Yaml: false}, // Yaml: false, returns JSON
			)

			// Setting kind manually
			d.Kind = "Deployment"

			// p.APIVersion = "v1"
			e.Encode(d.DeepCopyObject(), &deployRawJSON)
			deployJSON = deployRawJSON.String()
		}

		deploy = append(deploy, GetFormattedJSON(deployJSON))
	}
	return deploy, err
}

// GetClusterRoleBindingSpec provide a list an of JSON Cluster Role Binding specs from the clientset
func GetClusterRoleBindingSpec(cs kubernetes.Clientset, namespace string) (clusterRoleBinding []string, err error) {
	crb, err := cs.RbacV1().ClusterRoleBindings().List(metav1.ListOptions{})
	if err != nil {
		return
	}
	// fmt.Printf("crb: %v", crb)
	for _, c := range crb.Items {
		crbJSON := regexp.MustCompile(`\n|\|`).
			ReplaceAllString(c.Annotations["kubectl.kubernetes.io/last-applied-configuration"], "")

		if len(crbJSON) == 0 {
			var crbRawJSON strings.Builder
			e := k8sJSON.NewSerializerWithOptions(k8sJSON.DefaultMetaFactory,
				nil, nil,
				k8sJSON.SerializerOptions{Yaml: false}, // Yaml: false, returns JSON
			)

			// Setting kind manually
			c.Kind = "ClusterRoleBinding"

			// p.APIVersion = "v1"
			e.Encode(c.DeepCopyObject(), &crbRawJSON)
			crbJSON = crbRawJSON.String()
		}

		clusterRoleBinding = append(clusterRoleBinding, GetFormattedJSON(crbJSON))

	}
	return clusterRoleBinding, err
}

// GetIngressSpec provide a list an of JSON Ingress specs from the clientset
// /!\ The only rule we currently have doesn't seem to be triggerable
// /!\ K8S-IN-0001: https://github.com/reliablyhq/opa-policies/blob/main/kubernetes/ingress.rego
// /!\ It looks for indentical Ingress hosts in different namespaces, and we are currently working
// /!\ in only one namespace, passed as a parameter.
// /!\ Probably a TODO here.
func GetIngressSpec(cs kubernetes.Clientset, namespace string) (ingress []string, err error) {
	ing, err := cs.NetworkingV1beta1().Ingresses(namespace).List(metav1.ListOptions{})
	if err != nil {
		return
	}
	// fmt.Printf("crb: %v", crb)
	for _, i := range ing.Items {
		ingJSON := regexp.MustCompile(`\n|\|`).
			ReplaceAllString(i.Annotations["kubectl.kubernetes.io/last-applied-configuration"], "")

		if len(ingJSON) == 0 {
			var ingRawJSON strings.Builder
			e := k8sJSON.NewSerializerWithOptions(k8sJSON.DefaultMetaFactory,
				nil, nil,
				k8sJSON.SerializerOptions{Yaml: false}, // Yaml: false, returns JSON
			)

			// Setting kind manually
			i.Kind = "Ingress"

			// p.APIVersion = "v1"
			e.Encode(i.DeepCopyObject(), &ingRawJSON)
			// fmt.Printf("Ingress host: %v\n", ingRawJSON.Spec.rules.Host)
			ingJSON = ingRawJSON.String()
		}
		ingress = append(ingress, GetFormattedJSON(ingJSON))

	}
	return ingress, err
}

// GetPodSecurityPolicySpec provide a list an of JSON Pod Security Policy specs from the clientset
func GetPodSecurityPolicySpec(cs kubernetes.Clientset, namespace string) (podSecPol []string, err error) {
	secpol, err := cs.PolicyV1beta1().PodSecurityPolicies().List(metav1.ListOptions{})
	if err != nil {
		return
	}
	// fmt.Printf("crb: %v", crb)
	for _, p := range secpol.Items {
		secpolJSON := regexp.MustCompile(`\n|\|`).
			ReplaceAllString(p.Annotations["kubectl.kubernetes.io/last-applied-configuration"], "")

		if len(secpolJSON) == 0 {
			var secpolRawJSON strings.Builder
			e := k8sJSON.NewSerializerWithOptions(k8sJSON.DefaultMetaFactory,
				nil, nil,
				k8sJSON.SerializerOptions{Yaml: false}, // Yaml: false, returns JSON
			)

			// Setting kind manually
			p.Kind = "PodSecurityPolicy"

			// p.APIVersion = "v1"
			e.Encode(p.DeepCopyObject(), &secpolRawJSON)
			secpolJSON = secpolRawJSON.String()
		}

		podSecPol = append(podSecPol, GetFormattedJSON(secpolJSON))

	}
	return podSecPol, err
}
