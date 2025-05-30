package main

import (
	"context"
	"fmt"
	"os"
	"slices"
	"strings"

	"deepcover/src"

	"github.com/urfave/cli/v3"
)

func main() {
	app := &cli.Command{
		Name:        "deepcover",
		Usage:       "Identifies deep test coverage for dependencies",
		Description: "Analyzes test coverage starting from the specified entrypoint directory",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "run",
				Aliases: []string{"r"},
				Usage:   "Runs tests matching the provided regex",
				Value:   "",
			},
		},
		Action: run,
	}

	if err := app.Run(context.Background(), os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, cmd *cli.Command) error {
	entrypoint := cmd.Args().Get(0)
	if entrypoint == "" {
		return fmt.Errorf("entrypoint is required")
	}

	targetFunc := cmd.Args().Get(1)
	if targetFunc == "" {
		return fmt.Errorf("target function is required")
	}

	dependencies, err := src.GetDependencyFunctions(entrypoint, targetFunc)
	if err != nil {
		return fmt.Errorf("failed to get dependencies: %v", err)
	}

	expectedPackages := make([]string, 0)
	for _, dependency := range dependencies {
		if !slices.Contains(expectedPackages, dependency.PkgPath) {
			expectedPackages = append(expectedPackages, dependency.PkgPath)
		}
	}

	funcCoverages, err := src.GetCoverage(entrypoint, targetFunc, expectedPackages)
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
