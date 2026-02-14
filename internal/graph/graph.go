package graph

import "fmt"

type Node struct {
	ID      string
	Command string
	Line    int
}

type Edge struct {
	From string
	To   string
	Fail bool
	Line int
}

type Graph struct {
	Nodes map[string]*Node
	Edges []Edge
	// adjacency by success/fail
	NextSuccess map[string]string
	NextFail    map[string]string
	InDegree    map[string]int
}

func New() *Graph {
	return &Graph{
		Nodes:       map[string]*Node{},
		Edges:       []Edge{},
		NextSuccess: map[string]string{},
		NextFail:    map[string]string{},
		InDegree:    map[string]int{},
	}
}

func (g *Graph) AddNode(n Node) error {
	if _, ok := g.Nodes[n.ID]; ok {
		return fmt.Errorf("duplicate node id: %s (line %d)", n.ID, n.Line)
	}
	g.Nodes[n.ID] = &n
	if _, ok := g.InDegree[n.ID]; !ok {
		g.InDegree[n.ID] = 0
	}
	return nil
}

func (g *Graph) AddEdge(e Edge) error {
	if e.Fail {
		if _, exists := g.NextFail[e.From]; exists {
			return fmt.Errorf("duplicate fail edge from %s (line %d)", e.From, e.Line)
		}
		g.NextFail[e.From] = e.To
	} else {
		if _, exists := g.NextSuccess[e.From]; exists {
			return fmt.Errorf("duplicate success edge from %s (line %d)", e.From, e.Line)
		}
		g.NextSuccess[e.From] = e.To
	}
	g.Edges = append(g.Edges, e)
	g.InDegree[e.To]++
	if _, ok := g.InDegree[e.From]; !ok {
		g.InDegree[e.From] = 0
	}
	return nil
}

func (g *Graph) StartNode() (string, error) {
	var start string
	count := 0
	for id, deg := range g.InDegree {
		if deg == 0 {
			start = id
			count++
		}
	}
	if count == 0 {
		return "", fmt.Errorf("no start node (in-degree 0) found")
	}
	if count > 1 {
		return "", fmt.Errorf("multiple start nodes found: %d", count)
	}
	return start, nil
}
