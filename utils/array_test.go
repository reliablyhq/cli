package utils

import (
	"fmt"
	"testing"
)

func TestStringInArray(t *testing.T) {

	arr := []string{"a", "b", "c"}
	if !StringInArray("a", arr) {
		t.Error(fmt.Sprintf("'a' is missing from array %v", arr))
	}

	if StringInArray("x", arr) {
		t.Error(fmt.Sprintf("'x' shall not be in array %v", arr))
	}

}
