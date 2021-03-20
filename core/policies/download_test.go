package policies

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPolicyDownloader(t *testing.T) {

	tests := []struct {
		name       string
		policyName string
		want       string
		wantErr    error
	}{
		{
			name:       "valid policy",
			policyName: "kubernetes/v1/pod",
			want:       "this is a dummy policy content",
			wantErr:    nil,
		},
		{
			name:       "invalid policy",
			policyName: "kubernetes/fake/invalid",
			want:       "",
			wantErr:    ErrPolicyNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			pd := PolicyDownloader{}
			policy, err := pd.DownloadPolicy(tt.policyName)

			if tt.wantErr != nil {
				assert.Equal(t, ErrPolicyNotFound, err, "Expected error missing")
				return
			}

			assert.NoError(t, err, "Unexpected error")
			assert.NotEqual(t, "", string(policy), "Policy should not be empty")
		})
	}
}
