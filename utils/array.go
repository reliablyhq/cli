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
