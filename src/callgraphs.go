package src

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

func buildCallgraphs(path string, targetRegex *regexp.Regexp) (map[string]*callgraph.Graph, error) {
	ssaProg, ssaPkgs, err := buildSSA(chaConfig(), path)
	if err != nil {
		return nil, err
	}

	targetFuncs, err := findTargetSSAFunctions(ssaPkgs, targetRegex)
	if err != nil {
		return nil, err
	}

	cgs := make(map[string]*callgraph.Graph, len(targetFuncs))
	for _, target := range targetFuncs {
		cgs[target.Name()], err = buildCallgraph(cha.CallGraph, ssaProg, target)
		if err != nil {
			return nil, err
		}
	}

	return cgs, nil
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

func buildCallgraph(builder func(prog *ssa.Program) *callgraph.Graph, ssaProg *ssa.Program, target *ssa.Function) (*callgraph.Graph, error) {
	cg := builder(ssaProg)
	cg.DeleteSyntheticNodes()

	var ok bool
	if cg.Root, ok = cg.Nodes[target]; !ok {
		return nil, fmt.Errorf("failed to find callgraph node for function %s", target.Name())
	}

	return cg, nil
}
