package scanning

// Resource that can be scanned
type Resource struct {
	Actual   interface{}
	ID       string
	Platform string
}

// Result of a scan
type Result struct {
	Message string
}
