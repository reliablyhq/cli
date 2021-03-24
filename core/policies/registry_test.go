package policies

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type MockDownloader struct {
	Downloader
	cache map[string]string
}

func (md *MockDownloader) DownloadPolicy(id string) ([]byte, error) {
	if p, found := md.cache[id]; found {
		return []byte(p), nil
	}
	return []byte{}, ErrPolicyNotFound
}

func TestPolicyRegistry(t *testing.T) {

	pr := NewRegistry()
	pr.downloader = &MockDownloader{
		cache: map[string]string{
			"dummy":           "this is a dummy policy content",
			"a/custom/policy": "another one with key as path",
		},
	}

	tests := []struct {
		name       string
		policyName string
		want       string
		wantErr    error
	}{
		{
			name:       "basic",
			policyName: "dummy",
			want:       "this is a dummy policy content",
			wantErr:    nil,
		},
		{
			name:       "complex key id",
			policyName: "a/custom/policy",
			want:       "another one with key as path",
			wantErr:    nil,
		},
		{
			name:       "policy that does not exist",
			policyName: "invalid",
			want:       "",
			wantErr:    ErrPolicyNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			p, err := pr.GetPolicy(tt.policyName)

			if tt.wantErr != nil {
				assert.Equal(t, ErrPolicyNotFound, err, "Expected error missing")
				return
			}

			assert.NoError(t, err, "Unexpected error")
			assert.Equal(t, tt.want, string(p), "Unexpected policy content")
		})
	}

}

type SinglePolicyDownloader struct {
	Downloader
	cache map[string]string
}

func (spd *SinglePolicyDownloader) DownloadPolicy(id string) ([]byte, error) {
	if p, found := spd.cache[id]; found {
		// once we found it, we remove it to avoid consecutive download
		// in order to be able to check the caching system
		spd.cache[id] = "POLICY HAS ALREADY BEEN DOWNLOADED"

		return []byte(p), nil
	}
	return []byte{}, ErrPolicyNotFound
}

func TestPolicyRegistryCachedPolicy(t *testing.T) {

	// check cache is empty

	// fetch a policy

	// check policy is now stored in cache

	// fetch same policy a second time

	// check downloader has not been used a second time

	// check fetched policies are equal - and no error occured

	policyStr := "this is a dummy policy content"

	pr := NewRegistry()
	pr.downloader = &SinglePolicyDownloader{
		cache: map[string]string{
			"dummy": policyStr,
		},
	}

	// cache is empty at start
	pl, err := pr.registry.ListPolicies()
	assert.NoError(t, err, "Unexpected error")
	assert.Equal(t, []string{}, pl, "Cache should be empty")

	// fetch the policy from registry - use the downloader and put it into cache
	p1, err := pr.GetPolicy("dummy")
	assert.NoError(t, err, "Unexpected error")
	assert.Equal(t, policyStr, string(p1), "Wrong policy data")

	// cache is no more empty
	pl, err = pr.registry.ListPolicies()
	assert.NoError(t, err, "Unexpected error")
	assert.NotEqual(t, []string{}, pl, "Cache should not be empty")

	// Fetch the policy a second time from registry - should use cache
	// ie downloader should not be called
	p2, err := pr.GetPolicy("dummy")
	assert.NoError(t, err, "Unexpected error")
	assert.Equal(t, policyStr, string(p2), "Wrong policy data")

	// Both policy are equal - same content returned
	assert.Equal(t, string(p1), string(p2), "Policy data differ")
}
