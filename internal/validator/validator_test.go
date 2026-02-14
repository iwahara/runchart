package validator

import (
	"runchart/internal/graph"
	"testing"
)

func TestValidateOK(t *testing.T) {
	g := graph.New()
	_ = g.AddNode(graph.Node{ID: "A"})
	_ = g.AddNode(graph.Node{ID: "B"})
	_ = g.AddEdge(graph.Edge{From: "A", To: "B"})
	if err := Validate(g); err != nil {
		t.Fatalf("unexpected: %v", err)
	}
}

func TestValidateCycle(t *testing.T) {
	g := graph.New()
	_ = g.AddNode(graph.Node{ID: "A"})
	_ = g.AddNode(graph.Node{ID: "B"})
	_ = g.AddEdge(graph.Edge{From: "A", To: "B"})
	_ = g.AddEdge(graph.Edge{From: "B", To: "A"})
	if err := Validate(g); err == nil {
		t.Fatalf("expected cycle error")
	}
}

func TestValidateMultipleStarts(t *testing.T) {
	g := graph.New()
	_ = g.AddNode(graph.Node{ID: "A"})
	_ = g.AddNode(graph.Node{ID: "B"})
	if err := Validate(g); err == nil {
		t.Fatalf("expected multiple start nodes error")
	}
}
