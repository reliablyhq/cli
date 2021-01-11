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