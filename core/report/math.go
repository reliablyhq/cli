package report

import (
	"math"
)

func sum(array []float64) float64 {
	var f float64 = 0

	for _, x := range array {
		f += x
	}

	return f
}

func average(array []float64) float64 {
	l := len(array)
	if l == 0 {
		return 0
	}

	return sum(array) / float64(l)
}

func round2digits(f float64) float64 {
	return math.Round(f*100) / 100
}
