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

func TestTruncateStringTo20(t *testing.T) {
	const s = `Lorem ipsum dolor sit amet, consectetur adipiscing elit.
	Pellentesque faucibus a ante at viverra.`
	truncStr := TruncateString(s, 20)
	if len(truncStr) != 20 {
		t.Error("Truncated string does not have expected length")
	}
	if truncStr != "Lorem ipsum dolor..." {
		t.Error("Truncated string match the epxected string")
	}
	t.Logf("Truncated string:\n%s", truncStr)

}

func TestTruncateStringToDoesNotTruncate(t *testing.T) {
	const s = `Lorem ipsum dolor sit amet, consectetur adipiscing elit.
	Pellentesque faucibus a ante at viverra.`
	truncStr := TruncateString(s, len(s))
	if len(truncStr) != len(s) {
		t.Error("Truncated string does not have expected length")
	}
	if truncStr != s {
		t.Error("Truncated string match the epxected string")
	}
	t.Logf("Truncated string length:\n%v", len(truncStr))
	t.Logf("Truncated string:\n%s", truncStr)

}
