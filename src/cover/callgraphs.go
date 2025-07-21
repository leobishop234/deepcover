package cover

import (
	"errors"
	"fmt"
	"go/token"
	"regexp"

	"golang.org/x/tools/go/callgraph"
	"golang.org/x/tools/go/callgraph/cha"
	"golang.org/x/tools/go/packages"
	"golang.org/x/tools/go/ssa"
	"golang.org/x/tools/go/ssa/ssautil"
)

type callgraphAndTargets struct {
	callgraph *callgraph.Graph
	targets   []*callgraph.Node
}

func buildCallgraphs(path string, targetRegex *regexp.Regexp) (callgraphAndTargets, error) {
	ssaProg, ssaPkgs, err := buildSSA(chaConfig(), path)
	if err != nil {
		return callgraphAndTargets{}, err
	}

	targetFuncs, err := findTargetSSAFunctions(ssaPkgs, targetRegex)
	if err != nil {
		return callgraphAndTargets{}, err
	}

	results := callgraphAndTargets{
		callgraph: cha.CallGraph(ssaProg),
	}

	for _, target := range targetFuncs {
		targetNode, ok := results.callgraph.Nodes[target]
		if !ok {
			return callgraphAndTargets{}, fmt.Errorf("failed to find callgraph node for function %s", target.Name())
		}
		results.targets = append(results.targets, targetNode)
	}

	return results, nil
}

func chaConfig() *packages.Config {
	return &packages.Config{
		Mode:  packages.LoadSyntax | packages.NeedDeps | packages.NeedModule,
		Tests: true,
		Fset:  token.NewFileSet(),
	}
}

func buildSSA(conf *packages.Config, path string) (*ssa.Program, []*ssa.Package, error) {
	pkgs, err := packages.Load(conf, path)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load packages: %v", err)
	}
	if len(pkgs) == 0 {
		return nil, nil, fmt.Errorf("no packages found")
	}

	errs := []error{}
	for _, pkg := range pkgs {
		if len(pkg.Errors) > 0 {
			errs = append(errs, fmt.Errorf("failed to load package %s: %v", pkg.PkgPath, pkg.Errors))
		}
	}
	if len(errs) > 0 {
		return nil, nil, errors.Join(errs...)
	}

	ssaProg, ssaPkgs := ssautil.AllPackages(pkgs, 0)
	ssaProg.Build()

	return ssaProg, ssaPkgs, nil
}

func findTargetSSAFunctions(ssaPkgs []*ssa.Package, targetRegex *regexp.Regexp) ([]*ssa.Function, error) {
	var targetFuncs []*ssa.Function
	for _, ssaPkg := range ssaPkgs {
		for _, member := range ssaPkg.Members {
			if fn, ok := member.(*ssa.Function); ok {
				if targetRegex.MatchString(fn.Name()) {
					targetFuncs = append(targetFuncs, fn)
				}
			}
		}
	}

	return targetFuncs, nil
}
