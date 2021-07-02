package api

type QueryBody struct {
	Kind string `json:"kind"`
	Limit int `json:"limit"`
	Labels map[string]string `json:"Labels"`
	forEach ForEach `json:"forEach"`
}

type ForEach struct {
	ObjectiveResult struct `json:"objectiveResult"`
}

type ObjectiveResult struct {
	Include bool `json:"include"`
}