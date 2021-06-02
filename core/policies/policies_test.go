package policies

/*
func TestDownloadPodPolicy(t *testing.T) {

	dir, _ := os.MkdirTemp(os.TempDir(), ".reliably")
	defer os.RemoveAll(dir)

	fpath, err := downloadPolicyToCache(dir, "Kubernetes", "Pod")

	if fpath == "" ||
		err != nil {
		t.Error("Couldn't download Pod policy")
	}

	t.Log(fpath)
}
*/

/*
func TestDownloadInvalidPolicy(t *testing.T) {

	dir, _ := os.MkdirTemp(os.TempDir(), ".reliably")
	defer os.RemoveAll(dir)

	fpath, err := downloadPolicyToCache(dir, "Kubernetes", "azertyuiop")
	if fpath != "" ||
		err == nil {
		t.Error("How come we succeeded to download an invalid policy ?!")
	}

}
*/
/*
func TestFetchCachedPolicy(t *testing.T) {

	dir, _ := os.MkdirTemp(os.TempDir(), ".reliably")
	defer os.RemoveAll(dir)

	//First call, places policy in cache; second call uses cached policy
	_, _ = FetchPolicy(dir, "Kubernetes", "Pod")
	cached, _ := FetchPolicy(dir, "Kubernetes", "Pod")

	t.Log(cached)

	if cached == "" {
		t.Error("Policy was not retrieved properly")
	}

}
*/
