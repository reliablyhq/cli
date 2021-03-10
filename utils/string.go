package utils

import (
	"crypto/rand"
	"encoding/hex"
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
// elipsis (...)
func TruncateString(s string, num int) string {
	truncStr := s
	if len(s) > num {
		if num > 3 {
			num -= 3
		}
		truncStr = s[0:num] + "..."
	}
	return truncStr
}
