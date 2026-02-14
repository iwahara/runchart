package validator

import (
	"fmt"
	"runchart/internal/graph"
)

// Validate performs static checks on the graph.
func Validate(g *graph.Graph) error {
	// undefined node references
	for _, e := range g.Edges {
		if _, ok := g.Nodes[e.From]; !ok {
			return fmt.Errorf("undefined node '%s' referenced at line %d", e.From, e.Line)
		}
		if _, ok := g.Nodes[e.To]; !ok {
			return fmt.Errorf("undefined node '%s' referenced at line %d", e.To, e.Line)
		}
	}

	// start node count
	if _, err := g.StartNode(); err != nil {
		return err
	}

	// cycle detection (DFS colors)
	color := map[string]int{} // 0=white,1=gray,2=black
	var visit func(string) error
	visit = func(u string) error {
		if color[u] == 1 {
			return fmt.Errorf("cycle detected involving node '%s'", u)
		}
		if color[u] == 2 {
			return nil
		}
		color[u] = 1
		if v, ok := g.NextSuccess[u]; ok {
			if err := visit(v); err != nil {
				return err
			}
		}
		if v, ok := g.NextFail[u]; ok {
			if err := visit(v); err != nil {
				return err
			}
		}
		color[u] = 2
		return nil
	}
	for id := range g.Nodes {
		if color[id] == 0 {
			if err := visit(id); err != nil {
				return err
			}
		}
	}
	return nil
}
