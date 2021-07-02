package entities

import "encoding/json"

type Node struct {
	ID       string   `json:"id"`
	Metadata Metadata `json:"metadata"`
	Kind     string   `json:"kind"`
}

type Edge struct {
	Source string `json:"source"`
	Target string `json:"target"`
}

type NodeGraph struct {
	Nodes []*Node `json:"nodes"`
	Edges []*Edge `json:"edges"`
}

/*
MarshalJSON - custom JSON unmarshaller for NodeGraph type
used to insure that empty values are returned  instead of null

Eg:
	{ "nodes": [], "edges": [] }
	vs.
	{ "nodes": null, "edges": null }
*/
func (g *NodeGraph) MarshalJSON() ([]byte, error) {
	if g.Edges == nil {
		g.Edges = []*Edge{}
	}

	if g.Nodes == nil {
		g.Nodes = []*Node{}
	}

	// type alias used to prevent recursive call
	// to json.Marshal for the same type
	type alias NodeGraph
	return json.Marshal(alias(*g))
}
