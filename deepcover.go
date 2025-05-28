package main

import (
	"context"
	"fmt"
	"os"
	"slices"

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

	for _, funcCoverage := range funcCoverages {
		fmt.Printf("%s | %s | %f\n", funcCoverage.Path, funcCoverage.Name, funcCoverage.Coverage)
	}

	return nil
}
