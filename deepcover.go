package main

import (
	"context"
	"fmt"
	"os"

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
		Action: func(ctx context.Context, cmd *cli.Command) error {
			entrypoint := cmd.Args().Get(0)
			if entrypoint == "" {
				return fmt.Errorf("entrypoint is required")
			}

			targetFunc := cmd.Args().Get(1)
			if targetFunc == "" {
				return fmt.Errorf("target function is required")
			}

			dependencies, err := GetDependencyFunctions(entrypoint, targetFunc)
			if err != nil {
				return fmt.Errorf("failed to get dependencies: %v", err)
			}

			for _, dependency := range dependencies {
				fmt.Printf("%s/%s/%s\n", dependency.moduleName, dependency.pkgName, dependency.funcName)
			}

			return nil
		},
	}

	if err := app.Run(context.Background(), os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
