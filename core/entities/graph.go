package entities

type Node struct {
	ID     string    `json:"id"`
	Entity Objective `json:"entity"`
	Kind   string    `json:"kind"`
}

type Edge struct {
	Source string `json:"source"`
	Target string `json:"target"`
}

type NodeGraph struct {
	Nodes []*Node `json:"nodes"`
	Edges []*Edge `json:"edges"`
}
