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

func TestPolicyRegistryCachedPolicy(t *testing.T) {

	// check cache is empty

	// fetch a policy

	// check policy is now stored in cache

	// fetch same policy a second time

	// check downloader has not been used a second time

	// check fetched policies are equal - and no error occured

}
