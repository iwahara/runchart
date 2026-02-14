package executor

import (
	"bytes"
	"context"
	"testing"

	"runchart/internal/graph"
)

type mockRunner struct {
	seq []int
	i   int
}

func (m *mockRunner) RunCommand(ctx context.Context, command string) (int, error) {
	if m.i >= len(m.seq) {
		return 0, nil
	}
	v := m.seq[m.i]
	m.i++
	return v, nil
}

func TestExecuteSuccessPath(t *testing.T) {
	g := graph.New()
	_ = g.AddNode(graph.Node{ID: "A", Command: "echo A"})
	_ = g.AddNode(graph.Node{ID: "B", Command: "echo B"})
	_ = g.AddEdge(graph.Edge{From: "A", To: "B"})
	var out bytes.Buffer
	ex := New(g, &mockRunner{seq: []int{0, 0}}, &out)
	code, err := ex.Execute(context.Background())
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if code != 0 {
		t.Fatalf("want 0 got %d", code)
	}
	s := out.String()
	if !containsAll(s, []string{"✔ A", "✔ B"}) {
		t.Fatalf("unexpected out: %s", s)
	}
}

func TestExecuteFailBranch(t *testing.T) {
	g := graph.New()
	_ = g.AddNode(graph.Node{ID: "A", Command: "exit 1"})
	_ = g.AddNode(graph.Node{ID: "C", Command: "echo C"})
	_ = g.AddEdge(graph.Edge{From: "A", To: "C", Fail: true})
	var out bytes.Buffer
	ex := New(g, &mockRunner{seq: []int{1, 0}}, &out)
	code, err := ex.Execute(context.Background())
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if code != 0 {
		t.Fatalf("want 0 got %d", code)
	}
	s := out.String()
	if !containsAll(s, []string{"✖ A", "→ branching to C", "✔ C"}) {
		t.Fatalf("unexpected out: %s", s)
	}
}

func TestExecuteFailNoBranch(t *testing.T) {
	g := graph.New()
	_ = g.AddNode(graph.Node{ID: "A", Command: "exit 2"})
	var out bytes.Buffer
	ex := New(g, &mockRunner{seq: []int{2}}, &out)
	code, err := ex.Execute(context.Background())
	if err == nil {
		t.Fatalf("expected error")
	}
	if code != 2 {
		t.Fatalf("want 2 got %d", code)
	}
}

func containsAll(s string, subs []string) bool {
	for _, sub := range subs {
		if !bytes.Contains([]byte(s), []byte(sub)) {
			return false
		}
	}
	return true
}
