package report

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
