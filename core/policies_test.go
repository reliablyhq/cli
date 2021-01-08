package core

import (
	"io/ioutil"
	"os"
	"testing"
)

func TestDownloadPodPolicy(t *testing.T) {

	dir, err := ioutil.TempDir(os.TempDir(), ".reliably")
	defer os.RemoveAll(dir)

	fpath, err := DownloadPolicyToCache(dir, "Kubernetes", "Pod")

	if fpath == "" ||
		err != nil {
		t.Error("Couldn't download Pod policy")
	}

	t.Log(fpath)
}

func TestDownloadInvalidPolicy(t *testing.T) {

	dir, err := ioutil.TempDir(os.TempDir(), ".reliably")
	defer os.RemoveAll(dir)

	fpath, err := DownloadPolicyToCache(dir, "Kubernetes", "azertyuiop")
	if fpath != "" ||
		err == nil {
		t.Error("How come we succeeded to download an invalid policy ?!")
	}

}

func TestFetchCachedPolicy(t *testing.T) {

	dir, _ := ioutil.TempDir(os.TempDir(), ".reliably")
	defer os.RemoveAll(dir)

	//First call, places policy in cache; second call uses cached policy
	policy, _ := FetchPolicy(dir, "Kubernetes", "Pod")
	policy, _ = FetchPolicy(dir, "Kubernetes", "Pod")

	t.Log(policy)

	if policy == "" {
		t.Error("Policy was not retrieved properly")
	}

}
