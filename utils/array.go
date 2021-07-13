package utils

import (
	"reflect"
)

// StringInArray indicates whether a string value is contained in an array or slice
func StringInArray(str string, list []string) bool {
	for _, val := range list {
		if val == str {
			return true
		}
	}
	return false
}

// https://gist.githubusercontent.com/habibridho/e54ee3cb4d2e87708875cda75827a5bb/raw/ddcf3f3fd20665dd966161afdf93cc2e1127bb20/generic-filter.go
func Filter(arr interface{}, cond func(interface{}) bool) interface{} {
	contentType := reflect.TypeOf(arr)
	contentValue := reflect.ValueOf(arr)

	newContent := reflect.MakeSlice(contentType, 0, 0)
	for i := 0; i < contentValue.Len(); i++ {
		if content := contentValue.Index(i); cond(content.Interface()) {
			newContent = reflect.Append(newContent, content)
		}
	}
	return newContent.Interface()
}

// SumInt returns the sum of all int values of the slice
func SumInt(array []int) int {
	result := 0
	for _, v := range array {
		result += v
	}
	return result
}

// SumFloat64 returns the sum of all float64 values of the slice
func SumFloat64(array []float64) float64 {
	result := 0.0
	for _, v := range array {
		result += v
	}
	return result
}

// Reverse a slice
func Reverse(s interface{}) {
	n := reflect.ValueOf(s).Len()
	swap := reflect.Swapper(s)
	for i, j := 0, n-1; i < j; i, j = i+1, j-1 {
		swap(i, j)
	}
}

// AvgInt returns the average of all int values of the slice
func AvgInt(array []int) float64 {
	return float64(SumInt(array) / len(array))
}

// SumFloat64 returns the average of all float64 values of the slice
func AvgFloat64(array []float64) float64 {
	return SumFloat64(array) / float64(len(array))
}
