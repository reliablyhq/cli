package utils

import (
	"crypto/rand"
	"encoding/hex"
	"unicode/utf8"

	"github.com/acarl005/stripansi"
)

// RandomString generates a random string of a given length
func RandomString(length int) (string, error) {
	b := make([]byte, length/2)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// ansiDiff retruns the difference in length between the string in runes and
// the length of the string - if the strings have special characters the
// lengths are different
func runeDiff(s string) int {
	runeCount := utf8.RuneCountInString(s)
	strLen := len(s)
	return strLen - runeCount
}

// ansiDiff returns the difference in length between the string and the  same
// string when it is stipped of non ansi characters
func ansiDiff(s string) int {
	ansiStr := stripansi.Strip(s)
	return len(s) - len(ansiStr)
}

// TruncateString truncates the given string to the num provided and adds an
// ellipsis (...).
func TruncateString(s string, num int) string {
	truncStr := s
	nonPrintLen := runeDiff(s) + ansiDiff(s)
	printLen := len(s) - nonPrintLen

	if printLen > num {
		if num > 3 {
			num -= 3
			truncStr = s[0:num+nonPrintLen] + "..."
		} else {
			truncStr = s[0 : num+nonPrintLen]
		}
		// +nonPrintLen accounts for the  difference between the printable and non
		// printable characters

	}
	return truncStr
}
