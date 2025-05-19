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

func generateCallgraph(path, rootFunction string) (*callgraph.Graph, error) {
	cfg := &packages.Config{
		Mode:  packages.LoadSyntax | packages.NeedDeps,
		Tests: true,
		Fset:  token.NewFileSet(),
	}

	pkgs, err := packages.Load(cfg, path)
	if err != nil {
		return nil, fmt.Errorf("failed to load packages: %v", err)
	}

	ssaProg, ssaPkgs := ssautil.AllPackages(pkgs, 0)
	ssaProg.Build()

	var targetFunc *ssa.Function
	for _, pkg := range ssaPkgs {
		for _, member := range pkg.Members {
			if fn, ok := member.(*ssa.Function); ok && fn.Name() == rootFunction {
				targetFunc = fn
				break
			}
		}
	}

	if targetFunc == nil {
		return nil, fmt.Errorf("target function not found in %s", path)
	}

	cg := cha.CallGraph(ssaProg)
	cg.DeleteSyntheticNodes()

	cg.Root = cg.Nodes[targetFunc]
	if cg.Root == nil {
		return nil, fmt.Errorf("failed to find callgraph node for function %s", rootFunction)
	}

	return cg, nil
}

type knownPackage struct {
	moduled bool
	module  string
}

var knownPackages = map[string]knownPackage{}

func getModule(node *callgraph.Node) (string, bool, error) {
	if node.Func == nil || node.Func.Pkg == nil || node.Func.Pkg.Pkg == nil {
		return "", false, nil
	}

	pkg := node.Func.Pkg.Pkg.Path()

	if known, ok := knownPackages[pkg]; ok {
		return known.module, known.moduled, nil
	}

	pkgs, err := packages.Load(&packages.Config{
		Mode: packages.NeedName | packages.NeedModule,
	}, pkg)
	if err != nil {
		return "", false, err
	}

	if len(pkgs) == 0 || pkgs[0].Module == nil {
		knownPackages[pkg] = knownPackage{
			moduled: false,
			module:  "",
		}
	} else {
		knownPackages[pkg] = knownPackage{
			moduled: true,
			module:  pkgs[0].Module.Path,
		}
	}

	return knownPackages[pkg].module, knownPackages[pkg].moduled, nil
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

			for _, node := range cg.Nodes {
				module, ok, err := getModule(node)
				if err != nil {
					return fmt.Errorf("failed to get module: %v", err)
				}
				if ok {
					fmt.Println(module)
				}
			}

			return nil
		},
	}

	if err := app.Run(context.Background(), os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
