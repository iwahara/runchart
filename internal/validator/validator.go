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

	// v0.2: cycles are allowed. Instead, detect unreachable nodes from the start node.
	start, err := g.StartNode()
	if err != nil {
		return err
	}
	seen := map[string]bool{}
	var dfs func(string)
	dfs = func(u string) {
		if seen[u] {
			return
		}
		seen[u] = true
		if v, ok := g.NextSuccess[u]; ok {
			dfs(v)
		}
		if v, ok := g.NextFail[u]; ok {
			dfs(v)
		}
		if v, ok := g.NextDefault[u]; ok {
			dfs(v)
		}
		if m, ok := g.NextByCode[u]; ok {
			for _, v := range m {
				dfs(v)
			}
		}
	}
	dfs(start)
	for id := range g.Nodes {
		if !seen[id] {
			return fmt.Errorf("unreachable node '%s' from start '%s'", id, start)
		}
	}
	return nil
}
