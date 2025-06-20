package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"deepcover/src"
)

func main() {
	flag.Parse()
	args := flag.Args()
	if len(args) != 2 {
		fmt.Fprintf(os.Stderr, "Error: Expected 2 arguments (entrypoint and target function)\n")
		os.Exit(1)
	}

	entrypoint := args[0]
	targetFunc := args[1]

	if err := run(entrypoint, targetFunc); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run(entrypoint, targetFunc string) error {
	if entrypoint == "" {
		return fmt.Errorf("entrypoint is required")
	}

	if targetFunc == "" {
		return fmt.Errorf("target function is required")
	}

	dependencies, err := src.GetDependencies(entrypoint, targetFunc)
	if err != nil {
		return fmt.Errorf("failed to get dependencies: %v", err)
	}

	funcCoverages, err := src.GetCoverage(entrypoint, targetFunc, dependencies)
	if err != nil {
		return fmt.Errorf("failed to get coverage: %v", err)
	}

	displayCoverage(funcCoverages)

	return nil
}

func displayCoverage(funcCoverages []src.FunctionCoverage) {
	var pathLen, nameLen, coverageLen int

	for _, funcCoverage := range funcCoverages {
		if len(funcCoverage.Path) > pathLen {
			pathLen = len(funcCoverage.Path)
		}
		if len(funcCoverage.Name) > nameLen {
			nameLen = len(funcCoverage.Name)
		}
		if len(fmt.Sprintf("%.1f%%", funcCoverage.Coverage)) > coverageLen {
			coverageLen = len(fmt.Sprintf("%.1f%%", funcCoverage.Coverage))
		}
	}
	pathLen += 2
	nameLen += 2
	coverageLen += 2

	title := fmt.Sprintf("%-*s %-*s %-*s", pathLen, "PATH", nameLen, "FUNCTION", coverageLen, "COVERAGE")
	fmt.Println(title)
	fmt.Println(strings.Repeat("-", len(title)))

	for _, funcCoverage := range funcCoverages {
		coverageStr := fmt.Sprintf("%.1f%%", funcCoverage.Coverage)
		fmt.Printf("%-*s %-*s %-*s\n",
			pathLen,
			truncateString(funcCoverage.Path, pathLen),
			nameLen,
			truncateString(funcCoverage.Name, nameLen),
			coverageLen,
			coverageStr)
	}
	fmt.Println()
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
