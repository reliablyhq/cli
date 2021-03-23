package policies

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStorage(t *testing.T) {

	dir, err := ioutil.TempDir("", "storage_")
	if err != nil {
		t.Error("Unable to create temporary folder")
	}
	defer os.RemoveAll(dir)
	assert.NotEqual(t, "", dir)

	tests := []struct {
		name    string
		storage Store
	}{
		{
			name:    "in-memory store",
			storage: NewMemStore(),
		},
		{
			name:    "file system store",
			storage: NewFSStore(dir),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			s := tt.storage

			policies, err := s.ListPolicies()
			assert.NoError(t, err, "Unexpected error")
			assert.Equal(t, []string{}, policies, "Store should be empty!")

			e1 := s.UpsertPolicy("123", []byte("this is a policy"))
			e2 := s.UpsertPolicy("a/456", []byte("this is another one"))
			e3 := s.UpsertPolicy("b/c/789", []byte("and yet a third !!"))
			assert.ElementsMatch(t, []error{nil, nil, nil}, []error{e1, e2, e3})

			policies, err = s.ListPolicies()
			assert.NoError(t, err, "Unexpected error")
			assert.ElementsMatch(t, []string{"123", "a/456", "b/c/789"}, policies, "Store should not be empty!")

			p, err := s.GetPolicy("123")
			assert.NoError(t, err, "Unexpected error")
			assert.Equal(t, []byte("this is a policy"), p, "Unexpected policy data")

			_, err = s.GetPolicy("000")
			assert.Error(t, err, "Expected error not returned")
			assert.EqualError(t, err, "Policy not found")

			exists := s.HasPolicy("123")
			assert.True(t, exists, "Policy should exist")
			exists = s.HasPolicy("000")
			assert.False(t, exists, "Policy should not exist")

		})
	}

}
