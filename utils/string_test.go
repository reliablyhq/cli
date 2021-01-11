package utils

import (
	"testing"
)

func TestRandomString(t *testing.T) {

	randStr, err := RandomString(20)
	if err != nil {
		t.Error("Couldn't generate a random string")
	}

	if len(randStr) != 20 {
		t.Error("Random string does not have expected length")
	}

	t.Logf("Random string:\n%s", randStr)

}
