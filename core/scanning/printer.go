package scanning

import "log"

// Print results
func Print(r ...*Result) {
	for _, x := range r {
		log.Println(x)
	}
}
