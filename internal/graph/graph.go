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
	// Legacy v0.1 flag: fail-branch. When true and no other label fields are set,
	// this represents a generic non-zero exit branch.
	Fail bool
	// v0.2 labels
	// If Code is not nil, this edge is chosen when exit code matches exactly.
	Code *int
	// Default edge taken when no other rule matches.
	Default bool
	Line    int
}

type Graph struct {
	Nodes map[string]*Node
	Edges []Edge
	// adjacency by success/fail
	NextSuccess map[string]string         // success (exit 0) edge
	NextFail    map[string]string         // generic fail (non-zero) edge
	NextDefault map[string]string         // default edge when nothing else matches
	NextByCode  map[string]map[int]string // exact exit code match edges
	InDegree    map[string]int
}

func New() *Graph {
	return &Graph{
		Nodes:       map[string]*Node{},
		Edges:       []Edge{},
		NextSuccess: map[string]string{},
		NextFail:    map[string]string{},
		NextDefault: map[string]string{},
		NextByCode:  map[string]map[int]string{},
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
	// Decide edge bucket by label fields (priority: Code, Success/Fail(Default false & Fail flag), Default)
	switch {
	case e.Code != nil:
		m, ok := g.NextByCode[e.From]
		if !ok {
			m = map[int]string{}
			g.NextByCode[e.From] = m
		}
		if _, exists := m[*e.Code]; exists {
			return fmt.Errorf("duplicate exit-code edge from %s for code %d (line %d)", e.From, *e.Code, e.Line)
		}
		m[*e.Code] = e.To
	case e.Default:
		if _, exists := g.NextDefault[e.From]; exists {
			return fmt.Errorf("duplicate default edge from %s (line %d)", e.From, e.Line)
		}
		g.NextDefault[e.From] = e.To
	case e.Fail:
		if _, exists := g.NextFail[e.From]; exists {
			return fmt.Errorf("duplicate fail edge from %s (line %d)", e.From, e.Line)
		}
		g.NextFail[e.From] = e.To
	default:
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
