/*
package kubernetes is a colletion of funtions for interactng with a live
kubernetes cluster

*/
package kubernetes

import (
	"encoding/json"
	"log"
	"os"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

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
func GetPods(cs kubernetes.Clientset) ([]string, error) {
	var po []string
	namespace := "kube-system"
	pods, err := cs.CoreV1().Pods(namespace).List(metav1.ListOptions{})
	if err != nil {
		return po, err
	}
	for _, p := range pods.Items {
		name := p.GetName()
		pod, _ := cs.CoreV1().Pods(namespace).Get(name, metav1.GetOptions{})

		//MarshalIndent
		podJSON, err := json.MarshalIndent(pod, "", "  ")
		po = append(po, string(podJSON))
		if err != nil {
			log.Fatalf(err.Error())
		}
		// fmt.Printf(string(podJSON))
	}
	return po, nil
}
