package parser

import (
	"os"
	"path/filepath"
	"testing"
)

func writeTemp(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	p := filepath.Join(dir, "flow.mmd")
	if err := os.WriteFile(p, []byte(content), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}
	return p
}

func TestParseBasic(t *testing.T) {
	path := writeTemp(t, `flowchart TD
A[echo A]
A --> B
B[echo B]
`)
	res, err := Parse(path)
	if err != nil {
		t.Fatalf("parse err: %v", err)
	}
	if _, ok := res.Graph.Nodes["A"]; !ok {
		t.Fatalf("node A missing")
	}
	if _, ok := res.Graph.Nodes["B"]; !ok {
		t.Fatalf("node B missing")
	}
	if res.Graph.NextSuccess["A"] != "B" {
		t.Fatalf("edge A->B missing")
	}
}

func TestParseFailEdge(t *testing.T) {
	path := writeTemp(t, `flowchart TD
X[cmd]
Y[cmd]
X -- fail --> Y
`)
	res, err := Parse(path)
	if err != nil {
		t.Fatalf("parse err: %v", err)
	}
	if res.Graph.NextFail["X"] != "Y" {
		t.Fatalf("fail edge missing")
	}
}

func TestParseSyntaxError(t *testing.T) {
	path := writeTemp(t, `flowchart TD
bad line here
`)
	_, err := Parse(path)
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestParseLabeledEdgesV02(t *testing.T) {
	path := writeTemp(t, `flowchart TD
A[cmd]
B[cmd]
C[cmd]
D[cmd]
A -- 1 --> B
A -- default --> C
A -- fail --> D
`)
	res, err := Parse(path)
	if err != nil {
		t.Fatalf("parse err: %v", err)
	}
	// exact code edge
	if m := res.Graph.NextByCode["A"]; m == nil || m[1] != "B" {
		t.Fatalf("missing code=1 edge A->B")
	}
	if res.Graph.NextDefault["A"] != "C" {
		t.Fatalf("missing default edge A->C")
	}
	if res.Graph.NextFail["A"] != "D" {
		t.Fatalf("missing fail edge A->D")
	}
}
