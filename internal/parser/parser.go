package parser

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"runchart/internal/graph"
)

var (
	reFlowchart = regexp.MustCompile(`^\s*flowchart\b`)
	reNode      = regexp.MustCompile(`^\s*([A-Za-z0-9_\-]+)\s*\[(.+)\]\s*$`)
	// Unlabeled success edge (legacy & default success)
	reEdgeSucc = regexp.MustCompile(`^\s*([A-Za-z0-9_\-]+)\s*--?>\s*([A-Za-z0-9_\-]+)\s*$`)
	// Explicit labeled edges: -- <label> --> where label can be 'fail', 'default', or integer
	reEdgeLabeled = regexp.MustCompile(`^\s*([A-Za-z0-9_\-]+)\s*--\s*([A-Za-z0-9_\-]+)\s*-->\s*([A-Za-z0-9_\-]+)\s*$`)
)

type Result struct {
	Graph *graph.Graph
}

// Parse reads a Mermaid flowchart file and builds a graph.
func Parse(path string) (*Result, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	g := graph.New()
	s := bufio.NewScanner(f)
	line := 0
	seenFlowchart := false
	for s.Scan() {
		line++
		raw := s.Text()
		// Strip UTF-8 BOM if present at the very beginning of the file (common on Windows editors)
		if line == 1 {
			raw = strings.TrimPrefix(raw, "\uFEFF")
		}
		trimmed := strings.TrimSpace(raw)
		if trimmed == "" {
			continue
		}
		if strings.HasPrefix(trimmed, "%%") || strings.HasPrefix(trimmed, "//") { // Mermaid comment or line comment
			continue
		}
		if !seenFlowchart {
			if reFlowchart.MatchString(trimmed) {
				seenFlowchart = true
				continue
			} else {
				// Allow leading non-flowchart empty/comments only, else error with line
				return nil, fmt.Errorf("syntax error at line %d: expected 'flowchart' declaration", line)
			}
		}

		if m := reNode.FindStringSubmatch(trimmed); m != nil {
			id := m[1]
			cmd := m[2]
			if err := g.AddNode(graph.Node{ID: id, Command: cmd, Line: line}); err != nil {
				return nil, err
			}
			continue
		}

		// labeled edge first (to not conflict with plain -->)
		if m := reEdgeLabeled.FindStringSubmatch(trimmed); m != nil {
			from, label, to := m[1], strings.ToLower(m[2]), m[3]
			// special case: allow '-'/ '->' variations are covered by regex; label processed here
			switch label {
			case "fail":
				if err := g.AddEdge(graph.Edge{From: from, To: to, Fail: true, Line: line}); err != nil {
					return nil, err
				}
			case "default":
				if err := g.AddEdge(graph.Edge{From: from, To: to, Default: true, Line: line}); err != nil {
					return nil, err
				}
			default:
				// try parse integer
				if code, perr := strconv.Atoi(label); perr == nil {
					c := code
					if err := g.AddEdge(graph.Edge{From: from, To: to, Code: &c, Line: line}); err != nil {
						return nil, err
					}
				} else {
					return nil, fmt.Errorf("syntax error at line %d: unsupported edge label '%s'", line, label)
				}
			}
			continue
		}

		if m := reEdgeSucc.FindStringSubmatch(trimmed); m != nil {
			from, to := m[1], m[2]
			if err := g.AddEdge(graph.Edge{From: from, To: to, Fail: false, Line: line}); err != nil {
				return nil, err
			}
			continue
		}

		// ignore other Mermaid constructs for MVP (like direction, styles)
		// but if it looks like content and we can't parse it, raise syntax error
		if !strings.HasPrefix(trimmed, "classDef") && !strings.HasPrefix(trimmed, "style") {
			return nil, fmt.Errorf("syntax error at line %d: unsupported or invalid line", line)
		}
	}
	if err := s.Err(); err != nil {
		return nil, err
	}
	if !seenFlowchart {
		return nil, fmt.Errorf("syntax error: no 'flowchart' declaration found")
	}
	return &Result{Graph: g}, nil
}
