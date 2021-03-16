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

// TruncateString truncates the given string to the num provided and adds an
// elipsis (...). This function truncates will not be accurate if the string
// contains special characters e.g if you use ■, it has a length of 2
func TruncateString(s string, num int) string {
	truncStr := s
	ansiStr := stripansi.Strip(s)
	ansiStrLen := utf8.RuneCountInString(ansiStr)
	offset := utf8.RuneCountInString(s) - ansiStrLen
	// fmt.Printf("%v\n", getStringLen(s)-getStrippedStringLen(s))

	if ansiStrLen > num {
		if num > 3 {
			num -= 3
		}
		// +offest accounts for hte difference bettween
		truncStr = s[0:num+offset] + "..."
	}
	return truncStr
}
