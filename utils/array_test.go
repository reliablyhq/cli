package utils

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStringInArray(t *testing.T) {

	arr := []string{"a", "b", "c"}
	if !StringInArray("a", arr) {
		t.Error(fmt.Sprintf("'a' is missing from array %v", arr))
	}

	if StringInArray("x", arr) {
		t.Error(fmt.Sprintf("'x' shall not be in array %v", arr))
	}

}

type A struct{ B }
type B struct{ C int }

func TestFilterArray(t *testing.T) {
	tests := []struct {
		name  string
		array []interface{}
		wants []interface{}
		cond  func(interface{}) bool
	}{
		{
			name:  "empty array",
			array: []interface{}{},
			wants: []interface{}{},
			cond:  func(val interface{}) bool { return false },
		},
		{
			name:  "filter all",
			array: []interface{}{"a", "b", "c"},
			wants: []interface{}{},
			cond:  func(val interface{}) bool { return false },
		},
		{
			name:  "filter a string",
			array: []interface{}{"a", "b", "c"},
			wants: []interface{}{"a", "c"},
			cond:  func(val interface{}) bool { return val != "b" },
		},
		{
			name:  "filter even numbers (integer)",
			array: []interface{}{1, 2, 3, 4, 5},
			wants: []interface{}{1, 3, 5},
			cond:  func(val interface{}) bool { return val.(int)%2 != 0 },
		},
		{
			name:  "filter a value not in the array",
			array: []interface{}{"a", "b", "c"},
			wants: []interface{}{"a", "b", "c"},
			cond:  func(val interface{}) bool { return val != "x" },
		},
		{
			name:  "filter with a struct",
			array: []interface{}{A{B{C: 1}}, A{B{C: 2}}, A{B{C: 3}}},
			wants: []interface{}{A{B{C: 2}}, A{B{C: 3}}},
			cond:  func(val interface{}) bool { return val.(A).B.C != 1 },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filtered := Filter(tt.array, tt.cond)
			assert.Equal(t, tt.wants, filtered, "Filter did not worked as expected")
		})
	}
}

func TestSumInt(t *testing.T) {
	a := []int{1, 2, 3, 4, 5}

	sum := SumInt(a)
	assert.Equal(t, 15, sum, "Sum is not expected value")
}

func TestSumFloat64(t *testing.T) {
	a := []float64{1.0, 2.0, 3.0, 4.0, 5.5}

	sum := SumFloat64(a)
	assert.Equal(t, 15.5, sum, "Sum is not expected value")
}

func TestReverseSlice(t *testing.T) {

	tests := []struct {
		name  string
		slice interface{}
		wants interface{}
	}{
		{
			name:  "empty slice",
			slice: []int{},
			wants: []int{},
		},
		{
			name:  "reverse a slice of integer",
			slice: []int{1, 2, 3, 4, 5},
			wants: []int{5, 4, 3, 2, 1},
		},
		{
			name:  "reverse a slice of float",
			slice: []float64{1.0, 2.0, 3.0, 4.0, 5.0},
			wants: []float64{5.0, 4.0, 3.0, 2.0, 1.0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Reverse(tt.slice)
			assert.Equal(t, tt.wants, tt.slice, "Reversing slice did not work as expected")
		})
	}
}
