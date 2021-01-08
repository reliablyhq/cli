package utils

import (
	//"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIfThenElse(t *testing.T) {
	assert.Equal(t, IfThenElse(1 == 1, "Yes", false), "Yes")
	assert.Equal(t, IfThenElse(1 != 1, nil, 1), 1)
	assert.Equal(t, IfThenElse(1 < 2, nil, "No"), nil)
}
