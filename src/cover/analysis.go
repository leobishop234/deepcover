package cover

import (
	"errors"
	"fmt"
	"go/ast"
	"go/token"
	"regexp"

	"golang.org/x/tools/go/callgraph"
	"golang.org/x/tools/go/callgraph/cha"
	"golang.org/x/tools/go/packages"
	"golang.org/x/tools/go/ssa"
	"golang.org/x/tools/go/ssa/ssautil"
)

func buildAnalysis(path string, targetRegex *regexp.Regexp) (analysis, error) {
	pkgs, err := loadPackages(chaConfig(), path)
	if err != nil {
		return analysis{}, err
	}

	ssaProg, ssaPkgs, err := buildSSAObjects(pkgs)
	if err != nil {
		return analysis{}, err
	}

	targetSSAs := findTargetSSAFunctions(ssaPkgs, targetRegex)

	results := analysis{
		callgraph:   cha.CallGraph(ssaProg),
		asts:        buildASTMap(pkgs),
		targetNodes: make(map[functionID]*callgraph.Node, len(targetSSAs)),
	}

	for functionID, targetSSA := range targetSSAs {
		targetNode, ok := results.callgraph.Nodes[targetSSA]
		if !ok {
			return analysis{}, fmt.Errorf("failed to find callgraph node for function %s", targetSSA.Name())
		}
		results.targetNodes[functionID] = targetNode
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

func loadPackages(conf *packages.Config, path string) ([]*packages.Package, error) {
	pkgs, err := packages.Load(conf, path)
	if err != nil {
		return nil, fmt.Errorf("failed to load packages: %v", err)
	}

	if len(pkgs) == 0 {
		return nil, fmt.Errorf("no packages found")
	}

	errs := []error{}
	for _, pkg := range pkgs {
		if len(pkg.Errors) > 0 {
			errs = append(errs, fmt.Errorf("failed to load package %s: %v", pkg.PkgPath, pkg.Errors))
		}
	}
	if len(errs) > 0 {
		return nil, errors.Join(errs...)
	}

	return pkgs, nil
}

func buildSSAObjects(pkgs []*packages.Package) (*ssa.Program, []*ssa.Package, error) {
	ssaProg, ssaPkgs := ssautil.AllPackages(pkgs, 0)
	ssaProg.Build()

	return ssaProg, ssaPkgs, nil
}

func isInbuiltFunction(name string) bool {
	if name == "init" || (len(name) > 4 && name[:4] == "init" && name[4] == '#') {
		return true
	} else if name == "main" {
		return true
	}
	return false
}

func findTargetSSAFunctions(pkgs []*ssa.Package, targetRegex *regexp.Regexp) map[functionID]*ssa.Function {
	targetFuncs := make(map[functionID]*ssa.Function)
	for _, ssaPkg := range pkgs {
		for _, member := range ssaPkg.Members {
			if fn, ok := member.(*ssa.Function); ok {
				if isInbuiltFunction(fn.Name()) {
					continue
				}
				if targetRegex.MatchString(fn.Name()) {
					targetFuncs[functionID{pkgPath: ssaPkg.Pkg.Path(), funcName: fn.Name()}] = fn
				}
			}
		}
	}

	return targetFuncs
}

func buildASTMap(pkgs []*packages.Package) map[functionID]*ast.FuncDecl {
	astFuncs := make(map[functionID]*ast.FuncDecl)
	for _, pkg := range pkgs {
		for _, file := range pkg.Syntax {
			ast.Inspect(file, func(n ast.Node) bool {
				if fn, ok := n.(*ast.FuncDecl); ok && !isInbuiltFunction(fn.Name.Name) {
					astFuncs[functionID{pkgPath: pkg.PkgPath, funcName: fn.Name.Name}] = fn
				}
				return true
			})
		}
	}
	return astFuncs
}
