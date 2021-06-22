package entities

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

// custom marshaller to allow marshal indent
// func (n *NodeGraph) UnmarshalJSON(b []byte) error {
// 	fmt.Println(n)
// 	// b, err := json.MarshalIndent(n, "", "  ")
// 	var ng NodeGraph
// 	if err := json.Unmarshal(b, &ng); err != nil {
// 		return err
// 	}

// 	// make sure edges is not null
// 	n = &ng
// 	if n.Edges == nil {
// 		n.Edges = []*Edge{}
// 	}
// 	return nil
// }
