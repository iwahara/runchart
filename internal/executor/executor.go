package executor

import (
	"context"
	"fmt"
	"io"
	"os/exec"
	"runtime"
	"time"

	"runchart/internal/graph"
)

type Runner interface {
	RunCommand(ctx context.Context, command string) (exitCode int, err error)
}

type SystemRunner struct{}

func (SystemRunner) RunCommand(ctx context.Context, command string) (int, error) {
	var c *exec.Cmd
	if runtime.GOOS == "windows" {
		c = exec.CommandContext(ctx, "cmd", "/C", command)
	} else {
		c = exec.CommandContext(ctx, "/bin/sh", "-c", command)
	}
	if err := c.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return exitErr.ExitCode(), nil
		}
		return -1, err
	}
	return 0, nil
}

// Executor performs sequential execution following success/fail edges.
type Executor struct {
	G      *graph.Graph
	Runner Runner
	Out    io.Writer
}

func New(g *graph.Graph, r Runner, out io.Writer) *Executor {
	if r == nil {
		r = SystemRunner{}
	}
	return &Executor{G: g, Runner: r, Out: out}
}

// Execute runs from start node until no next node. Returns final exit code and error when control-flow error occurs.
func (e *Executor) Execute(ctx context.Context) (int, error) {
	start, err := e.G.StartNode()
	if err != nil {
		return -1, err
	}
	curr := start
	lastExit := 0
	visited := map[string]bool{}
	for curr != "" {
		n := e.G.Nodes[curr]
		if n == nil {
			return -1, fmt.Errorf("internal error: missing node %s", curr)
		}
		t0 := time.Now()
		code, runErr := e.Runner.RunCommand(ctx, n.Command)
		dur := time.Since(t0)
		lastExit = code
		if code == 0 {
			fmt.Fprintf(e.Out, "✔ %s (%.1fs)\n", n.ID, dur.Seconds())
			next := e.G.NextSuccess[n.ID]
			curr = next
		} else {
			fmt.Fprintf(e.Out, "✖ %s (exit %d)\n", n.ID, code)
			next, ok := e.G.NextFail[n.ID]
			if !ok {
				return code, fmt.Errorf("no fail branch defined for node '%s' (exit %d)", n.ID, code)
			}
			fmt.Fprintf(e.Out, "→ branching to %s\n", next)
			curr = next
		}
		// stop if we see self-loop or unexpected revisit (safety, though validator should prevent cycles)
		if visited[curr] {
			return lastExit, fmt.Errorf("cycle detected at runtime involving '%s'", curr)
		}
		visited[curr] = true
		if runErr != nil { // non-exit error
			return lastExit, runErr
		}
	}
	return lastExit, nil
}
