package cli

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"time"

	"runchart/internal/executor"
	"runchart/internal/parser"
	"runchart/internal/validator"
)

// Run parses, validates, and executes the given Mermaid file.
// Returns a process exit code.
func Run(path string, out io.Writer, errOut io.Writer, maxSteps int) int {
	res, err := parser.Parse(path)
	if err != nil {
		fmt.Fprintf(errOut, "%v\n", err)
		return 2
	}
	if err := validator.Validate(res.Graph); err != nil {
		fmt.Fprintf(errOut, "%v\n", err)
		return 2
	}

	ex := executor.New(res.Graph, nil, out)
	if maxSteps > 0 {
		ex.MaxSteps = maxSteps
	}
	// Cancel on timeout or SIGINT
	baseCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	ctx, cancel := context.WithTimeout(baseCtx, 24*time.Hour)
	defer cancel()
	code, runErr := ex.Execute(ctx)
	if runErr != nil {
		fmt.Fprintf(errOut, "%v\n", runErr)
		// if execution failed due to control flow error, propagate non-zero
		if code == 0 {
			return 1
		}
		return code
	}
	return code
}

// Validate parses and statically validates the given Mermaid file without executing it.
// Returns 0 when valid, or 2 when a parse/validate error occurs.
func Validate(path string, out io.Writer, errOut io.Writer) int {
	res, err := parser.Parse(path)
	if err != nil {
		fmt.Fprintf(errOut, "%v\n", err)
		return 2
	}
	if err := validator.Validate(res.Graph); err != nil {
		fmt.Fprintf(errOut, "%v\n", err)
		return 2
	}
	if out != nil {
		fmt.Fprintln(out, "valid")
	}
	return 0
}
