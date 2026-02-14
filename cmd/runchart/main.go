package main

import (
	"flag"
	"fmt"
	"os"

	"runchart/internal/cli"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}

	cmd := os.Args[1]
	switch cmd {
	case "run":
		fs := flag.NewFlagSet("run", flag.ExitOnError)
		_ = fs.Parse(os.Args[2:])
		args := fs.Args()
		if len(args) != 1 {
			fmt.Fprintln(os.Stderr, "usage: runchart run <flow.mmd>")
			os.Exit(2)
		}
		path := args[0]
		code := cli.Run(path, os.Stdout, os.Stderr)
		os.Exit(code)
	case "-h", "--help", "help":
		usage()
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", cmd)
		usage()
		os.Exit(2)
	}
}

func usage() {
	fmt.Println("runchart – execute Mermaid flowchart as control flow")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  run <flow.mmd>   Execute the flowchart")
}
