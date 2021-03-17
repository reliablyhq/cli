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
	expectedStr := "Lorem ipsum dolor..."
	truncStr := TruncateString(s, 20)
	if len(truncStr) != len(expectedStr) {
		t.Error("Truncated string does not have expected length")
	}
	if truncStr != expectedStr {
		t.Error("Truncated string match the epxected string")
	}
	t.Logf("Truncated string:\n%s", truncStr)

}

func TestTruncateShortStringDoesNotAddEllipse(t *testing.T) {
	const s = `Lorem ipsum dolor sit amet, consectetur adipiscing elit.
	Pellentesque faucibus a ante at viverra.`
	expectedStr := "Lor"
	truncStr := TruncateString(s, 3)
	if len(truncStr) != len(expectedStr) {
		t.Error("Truncated string does not have expected length")
	}
	if truncStr != expectedStr {
		t.Error("Truncated string match the epxected string")
	}
	t.Logf("Truncated string:\n%s", truncStr)

}

func TestDoesNotTruncateLessThan20(t *testing.T) {
	const s = `Lorem ipsum dolor si`
	expectedStr := "Lorem ipsum dolor si"
	truncStr := TruncateString(s, 20)
	if len(truncStr) != len(expectedStr) {
		t.Error("Truncated string does not have expected length")
	}
	if truncStr != expectedStr {
		t.Error("Truncated string match the epxected string")
	}
	t.Logf("Truncated string:\n%s", truncStr)

}

func TestTruncatesColoredStringWithSymbol(t *testing.T) {

	// define string in bytes as it has control codes and special characters ("■")
	ba := []byte{27, 91, 51, 51, 109, 226, 150, 160, 27, 91, 48, 109, 32, 32, 101, 120, 97, 109, 112, 108, 101, 58, 48, 58, 48, 101, 120, 97, 109, 112, 108, 101, 58, 48, 58, 48}

	sourceStr := string(ba)

	expectedString := string(ba[0:25]) + "..."

	truncStr := TruncateString(sourceStr, 17)
	if len(truncStr) != len(expectedString) {
		t.Error("Truncated string does not have expected length")
	}
	if truncStr != expectedString {
		t.Error("Truncated string match the epxected string")
	}
	t.Logf("Truncated string:\n%s", truncStr)

}

func TestTruncatesShortColoredStringWithSymbolDoesNotAddEllipse(t *testing.T) {

	// define string in bytes as it has control codes and special characters ("■")
	ba := []byte{27, 91, 51, 51, 109, 226, 150, 160, 27, 91, 48, 109, 32, 32, 101, 120, 97, 109, 112, 108, 101, 58, 48, 58, 48, 101, 120, 97, 109, 112, 108, 101, 58, 48, 58, 48}

	sourceStr := string(ba)

	expectedString := string(ba[0 : 11+3])

	truncStr := TruncateString(sourceStr, 3)
	if len(truncStr) != len(expectedString) {
		t.Error("Truncated string does not have expected length")
	}
	if truncStr != expectedString {
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
