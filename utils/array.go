package utils

// StringInArray indicates whether a string value is contained in an array or slice
func StringInArray(str string, list []string) bool {
	for _, val := range list {
		if val == str {
			return true
		}
	}
	return false
}
