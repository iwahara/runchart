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
	// MaxSteps limits the number of transitions to prevent infinite loops (v0.2)
	MaxSteps int
}

func New(g *graph.Graph, r Runner, out io.Writer) *Executor {
	if r == nil {
		r = SystemRunner{}
	}
	return &Executor{G: g, Runner: r, Out: out, MaxSteps: 1000}
}

// Execute runs from start node until no next node. Returns final exit code and error when control-flow error occurs.
func (e *Executor) Execute(ctx context.Context) (int, error) {
	start, err := e.G.StartNode()
	if err != nil {
		return -1, err
	}
	curr := start
	lastExit := 0
	if e.MaxSteps <= 0 {
		e.MaxSteps = 1000
	}
	steps := 0
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
		} else {
			fmt.Fprintf(e.Out, "✖ %s (exit %d)\n", n.ID, code)
		}

		// decide next node based on v0.2 priority
		var next string
		if m, ok := e.G.NextByCode[n.ID]; ok {
			if v, ok2 := m[code]; ok2 {
				next = v
			}
		}
		if next == "" {
			if code == 0 {
				if v, ok := e.G.NextSuccess[n.ID]; ok {
					next = v
				}
			} else {
				if v, ok := e.G.NextFail[n.ID]; ok {
					next = v
				}
			}
		}
		if next == "" {
			if v, ok := e.G.NextDefault[n.ID]; ok {
				next = v
			}
		}

		if next == "" { // end of flow
			break
		}
		if code != 0 { // keep legacy branching message for non-zero
			fmt.Fprintf(e.Out, "→ branching to %s\n", next)
		}
		if visited[next] {
			fmt.Fprintf(e.Out, "↺ loop to %s\n", next)
		}
		curr = next
		steps++
		if steps >= e.MaxSteps {
			return lastExit, fmt.Errorf("execution aborted: max steps (%d) exceeded", e.MaxSteps)
		}
		visited[curr] = true
		if runErr != nil { // non-exit error
			return lastExit, runErr
		}
	}
	return lastExit, nil
}
