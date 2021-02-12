package utils

import (
	//"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIfThenElse(t *testing.T) {
	//nolint:staticcheck // SA1004 ignore this!
	assert.Equal(t, "Yes", IfThenElse(1 == 1, "Yes", false))
	//nolint:staticcheck // SA1004 ignore this!
	assert.Equal(t, 1, IfThenElse(1 != 1, nil, 1))
	assert.Equal(t, nil, IfThenElse(1 < 2, nil, "No"))
}
