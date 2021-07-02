package utils

import (
	"math/rand"
	"time"
)

// RandomInt - returns an int >= min, < max
func RandomInt(min, max int64) int64 {
	rand.Seed(time.Now().UnixNano())
	return min + int64(rand.Int63n(max-min))
}
