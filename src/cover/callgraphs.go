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

type callgraphAndTargets struct {
	callgraph *callgraph.Graph
	targets   map[string]targetObj
}

type targetObj struct {
	ssaFunc *ssa.Function
	node    *callgraph.Node
	ast     *ast.FuncDecl
}

func buildCallgraphs(path string, targetRegex *regexp.Regexp) (callgraphAndTargets, error) {
	pkgs, err := loadPackages(chaConfig(), path)
	if err != nil {
		return callgraphAndTargets{}, err
	}

	ssaProg, ssaPkgs, err := buildObjects(pkgs)
	if err != nil {
		return callgraphAndTargets{}, err
	}

	targetSSAs, err := findTargetFunctions(ssaPkgs, targetRegex)
	if err != nil {
		return callgraphAndTargets{}, err
	}

	targetASTs, err := findTargetASTFunctions(pkgs, targetRegex)
	if err != nil {
		return callgraphAndTargets{}, err
	}

	results := callgraphAndTargets{
		callgraph: cha.CallGraph(ssaProg),
		targets:   make(map[string]targetObj, len(targetSSAs)),
	}

	for ssaName, targetSSA := range targetSSAs {
		targetNode, ok := results.callgraph.Nodes[targetSSA]
		if !ok {
			return callgraphAndTargets{}, fmt.Errorf("failed to find callgraph node for function %s", targetSSA.Name())
		}

		targetAST, ok := targetASTs[ssaName]
		if !ok {
			return callgraphAndTargets{}, fmt.Errorf("failed to find AST function for function %s", ssaName)
		}

		results.targets[ssaName] = targetObj{
			node:    targetNode,
			ssaFunc: targetSSA,
			ast:     targetAST,
		}
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

func buildObjects(pkgs []*packages.Package) (*ssa.Program, []*ssa.Package, error) {
	ssaProg, ssaPkgs := ssautil.AllPackages(pkgs, 0)
	ssaProg.Build()

	return ssaProg, ssaPkgs, nil
}

func isInitFunction(name string) bool {
	return name == "init" || (len(name) > 4 && name[:4] == "init" && name[4] == '#')
}

func findTargetFunctions(pkgs []*ssa.Package, targetRegex *regexp.Regexp) (map[string]*ssa.Function, error) {
	targetFuncs := make(map[string]*ssa.Function)
	for _, ssaPkg := range pkgs {
		for _, member := range ssaPkg.Members {
			if fn, ok := member.(*ssa.Function); ok {
				// Skip init functions - they are not supported
				if isInitFunction(fn.Name()) {
					continue
				}
				if targetRegex.MatchString(fn.Name()) {
					targetFuncs[fn.Name()] = fn
				}
			}
		}
	}

	return targetFuncs, nil
}

func findTargetASTFunctions(pkgs []*packages.Package, targetRegex *regexp.Regexp) (map[string]*ast.FuncDecl, error) {
	astFuncs := make(map[string]*ast.FuncDecl)
	for _, pkg := range pkgs {
		for _, file := range pkg.Syntax {
			ast.Inspect(file, func(n ast.Node) bool {
				if fn, ok := n.(*ast.FuncDecl); ok {
					// Skip init functions - they are not supported
					if isInitFunction(fn.Name.Name) {
						return true
					}
					if targetRegex.MatchString(fn.Name.Name) {
						astFuncs[fn.Name.Name] = fn
					}
				}
				return true
			})
		}
	}

	return astFuncs, nil
}
