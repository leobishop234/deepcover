package main

import (
	"context"
	"fmt"
	"go/token"
	"os"

	"github.com/urfave/cli/v3"
	"golang.org/x/tools/go/callgraph"
	"golang.org/x/tools/go/callgraph/cha"
	"golang.org/x/tools/go/packages"
	"golang.org/x/tools/go/ssa"
	"golang.org/x/tools/go/ssa/ssautil"
)

func generateCallgraph(pkgPath, funcName string) (*callgraph.Graph, error) {
	cfg := &packages.Config{
		Mode:  packages.LoadAllSyntax,
		Tests: true,
		Fset:  token.NewFileSet(),
	}

	pkgs, err := packages.Load(cfg, pkgPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load packages: %v", err)
	}

	ssaProg, ssaPkgs := ssautil.AllPackages(pkgs, 0)
	ssaProg.Build()

	var targetFunc *ssa.Function
	for _, pkg := range ssaPkgs {
		for _, member := range pkg.Members {
			if fn, ok := member.(*ssa.Function); ok && fn.Name() == funcName {
				targetFunc = fn
				break
			}
		}
	}

	if targetFunc == nil {
		return nil, fmt.Errorf("target function not found in %s", pkgPath)
	}

	cg := cha.CallGraph(ssaProg)
	cg.DeleteSyntheticNodes()

	cg.Root = cg.Nodes[targetFunc]
	if cg.Root == nil {
		return nil, fmt.Errorf("failed to find callgraph node for function %s", funcName)
	}

	return cg, nil
}

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

			cg, err := generateCallgraph(entrypoint, targetFunc)
			if err != nil {
				return fmt.Errorf("failed to generate callgraph: %v", err)
			}

			fmt.Println(cg.Root)

			return nil
		},
	}

	if err := app.Run(context.Background(), os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
