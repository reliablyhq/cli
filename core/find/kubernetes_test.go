package find

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
)

const (
	ns = `
apiVersion: v1
kind: Namespace
metadata:
  name: reliably`

	pod = `
apiVersion: v1
kind: Pod
metadata:
  name: chaostoolkit
  namespace: chaostoolkit
  labels:
    app: chaostoolkit
spec:
  restartPolicy: Never
  containers:
  - name: chaostoolkit
    image: chaostoolkit:latest`
)

func TestExtractKindFromManifest(t *testing.T) {
	kind := ExtractKindFromManifest(ns)
	t.Logf("Kind: %v", kind)
	if strings.ToLower(kind) != "namespace" {
		t.Error("Kind not extracted properly; was expecting 'Namespace'")
	}
}

func TestUnmarshallYaml(t *testing.T) {
	var m KubernetesAPI
	err := yaml.Unmarshal([]byte(ns), &m)
	if err != nil {
		t.Errorf("Could not unmarshal: %v \n%v", err, ns)
	}
	t.Log(m)
}

func TestK8sURI(t *testing.T) {
	var m KubernetesAPI
	_ = yaml.Unmarshal([]byte(ns), &m)
	uri := m.URI()
	assert.Equal(t, "/api/v1/namespace/reliably", uri)

	_ = yaml.Unmarshal([]byte(pod), &m)
	uri = m.URI()
	assert.Equal(t, "/api/v1/pod/chaostoolkit/chaostoolkit", uri)
}
