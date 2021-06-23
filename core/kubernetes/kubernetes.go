/*
package kubernetes is a colletion of funtions for interactng with a live
kubernetes cluster

*/
package kubernetes

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/mitchellh/go-homedir"
	log "github.com/sirupsen/logrus"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	netv1beta1 "k8s.io/api/networking/v1beta1"
	policyv1beta1 "k8s.io/api/policy/v1beta1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
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

// FileExists checks if a file exists and is not a directory before we
// try using it to prevent further errors.
func FileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

// FindKubeConfigPath checks will return a path for the kubernetes config file
func FindKubeConfigPath(path ...string) (string, error) {
	if len(path) > 0 {
		return path[0], nil
	}

	env := os.Getenv("KUBECONFIG")
	if env != "" {
		return env, nil
	}

	p, _ := homedir.Dir()
	p = p + "/.kube/config"
	return p, nil
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
func GetKubernetesClientSet(kubeconfigPath, context string) (*kubernetes.Clientset, error) {
	// Pull the config
	config, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeconfigPath},
		&clientcmd.ConfigOverrides{CurrentContext: context}).
		ClientConfig()
	// config, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		return nil, err
	}

	// Connect
	return kubernetes.NewForConfig(config)
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
func GetPodSpec(ctx context.Context, cs kubernetes.Clientset, namespace string) (po []string, err error) {
	pods, err := cs.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return
	}

	for _, p := range pods.Items {
		podJSON := itemToJSON(p, "Pod")
		po = append(po, podJSON)
	}
	return po, err
}

// GetDeploymentSpec provide a list an of JSON Deployment specs from the clientset
func GetDeploymentSpec(ctx context.Context, cs kubernetes.Clientset, namespace string) (deploy []string, err error) {
	deployment, err := cs.AppsV1().Deployments(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return
	}

	for _, d := range deployment.Items {
		deployJSON := itemToJSON(d, "Deployment")
		deploy = append(deploy, deployJSON)
	}
	return deploy, err
}

// GetClusterRoleBindingSpec provide a list an of JSON Cluster Role Binding specs from the clientset
func GetClusterRoleBindingSpec(ctx context.Context, cs kubernetes.Clientset) (clusterRoleBinding []string, err error) {
	crb, err := cs.RbacV1().ClusterRoleBindings().List(ctx, metav1.ListOptions{})
	if err != nil {
		return
	}

	for _, c := range crb.Items {
		crbJSON := itemToJSON(c, "ClusterRoleBinding")
		clusterRoleBinding = append(clusterRoleBinding, crbJSON)
	}
	return clusterRoleBinding, err
}

// GetIngressSpec provide a list an of JSON Ingress specs from the clientset
// /!\ The only rule we currently have doesn't seem to be triggerable
// /!\ K8S-IN-0001: https://github.com/reliablyhq/opa-policies/blob/main/kubernetes/ingress.rego
// /!\ It looks for identical Ingress hosts in different namespaces, and we are currently working
// /!\ in only one namespace, passed as a parameter.
// /!\ Probably a TODO here.
func GetIngressSpec(ctx context.Context, cs kubernetes.Clientset, namespace string) (ingress []string, err error) {
	ing, err := cs.NetworkingV1beta1().Ingresses(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return
	}

	for _, i := range ing.Items {
		ingJSON := itemToJSON(i, "Ingress")
		ingress = append(ingress, ingJSON)
	}
	return ingress, err
}

// GetPodSecurityPolicySpec provide a list an of JSON Pod Security Policy specs from the clientset
func GetPodSecurityPolicySpec(ctx context.Context, cs kubernetes.Clientset) (podSecPol []string, err error) {
	secpol, err := cs.PolicyV1beta1().PodSecurityPolicies().List(ctx, metav1.ListOptions{})
	if err != nil {
		return
	}
	for _, p := range secpol.Items {
		secpolJSON := itemToJSON(p, "PodSecurityPolicy")
		podSecPol = append(podSecPol, secpolJSON)
	}
	return podSecPol, err
}

// GetNodeSpec returns the list of JSON Nodes specs from the clientset
func GetNodeSpec(ctx context.Context, cs kubernetes.Clientset) (nodes []string, err error) {
	np, err := cs.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return
	}

	for _, n := range np.Items {
		nodeJSON := itemToJSON(n, "Node")
		nodes = append(nodes, nodeJSON)
	}
	return nodes, err
}

func itemToJSON(item interface{}, kind string) string {
	const lastconfigkey = "kubectl.kubernetes.io/last-applied-configuration"
	var lastConfig string
	var dc runtime.Object
	switch kind {
	case "Pod":
		i := item.(corev1.Pod)
		i.Kind = kind
		i.APIVersion = "v1"
		lastConfig = i.Annotations[lastconfigkey]
		dc = i.DeepCopyObject()
	case "Deployment":
		i := item.(appsv1.Deployment)
		i.Kind = kind
		i.APIVersion = "app/v1"
		lastConfig = i.Annotations[lastconfigkey]
		dc = i.DeepCopyObject()
	case "ClusterRoleBinding":
		i := item.(rbacv1.ClusterRoleBinding)
		i.Kind = kind
		i.APIVersion = "rbac.authorization.k8s.io/v1"
		lastConfig = i.Annotations[lastconfigkey]
		dc = i.DeepCopyObject()
	case "Ingress":
		i := item.(netv1beta1.Ingress)
		i.Kind = kind
		i.APIVersion = "networking.k8s.io/v1"
		lastConfig = i.Annotations[lastconfigkey]
		dc = i.DeepCopyObject()
	case "PodSecurityPolicy":
		i := item.(policyv1beta1.PodSecurityPolicy)
		i.Kind = kind
		i.APIVersion = "policy/v1beta1"
		lastConfig = i.Annotations[lastconfigkey]
		dc = i.DeepCopyObject()
	case "Node":
		i := item.(corev1.Node)
		i.Kind = kind
		i.APIVersion = "v1"
		lastConfig = i.Annotations[lastconfigkey]
		dc = i.DeepCopyObject()
	}

	JSON := regexp.MustCompile(`\n|\|`).
		ReplaceAllString(lastConfig, "")

	// if annotation was empty use k8s JSON serializer
	if len(JSON) == 0 {
		var rawJSON strings.Builder
		e := k8sJSON.NewSerializerWithOptions(k8sJSON.DefaultMetaFactory,
			nil, nil,
			k8sJSON.SerializerOptions{Yaml: false}, // Yaml: false, returns JSON
		)

		// p.APIVersion = "v1"
		e.Encode(dc, &rawJSON)
		JSON = rawJSON.String()

		// log.Debugf("Error processing pod: %v\n", p.Name)

	}
	return GetFormattedJSON(JSON)
}

func GetResourceList(ctx context.Context, cs kubernetes.Clientset, namespace string) []string {
	var rl []string = make([]string, 0, 0)

	podList, _ := GetPodSpec(ctx, cs, namespace)
	deploymentList, _ := GetDeploymentSpec(ctx, cs, namespace)
	clusterRoleBindingList, _ := GetClusterRoleBindingSpec(ctx, cs)
	ingressList, _ := GetIngressSpec(ctx, cs, namespace)
	podSecurityPolicyList, _ := GetPodSecurityPolicySpec(ctx, cs)
	nodeList, _ := GetNodeSpec(ctx, cs)

	lists := [][]string{
		podList,
		deploymentList,
		clusterRoleBindingList,
		ingressList,
		podSecurityPolicyList,
		nodeList,
	}

	for _, l := range lists {
		rl = append(rl, l...)
	}

	return rl
}
