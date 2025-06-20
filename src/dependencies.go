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

type function struct {
	ModuleName string
	PkgName    string
	PkgPath    string
	FuncName   string
}

func GetDependencies(path string, targetRegex *regexp.Regexp) (map[string][]function, error) {
	cgs, err := buildCallgraphs(path, targetRegex)
	if err != nil {
		return nil, err
	}

	results := make(map[string][]function, len(cgs))
	for target, cg := range cgs {
		dependencies, err := getDependencies(cg)
		if err != nil {
			return nil, err
		}
		results[target] = dependencies
	}

	return results, nil
}

func buildCallgraphs(path string, targetRegex *regexp.Regexp) (map[string]*callgraph.Graph, error) {
	ssaProg, ssaPkgs, err := buildSSA(chaConfig(), path)
	if err != nil {
		return nil, err
	}

	targetFuncs, err := findTargetSSAFunctions(ssaPkgs, targetRegex)
	if err != nil {
		return nil, err
	}

	return generateCallgraphs(ssaProg, cha.CallGraph, targetFuncs)
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

func generateCallgraphs(ssaProg *ssa.Program, builder func(prog *ssa.Program) *callgraph.Graph, targets []*ssa.Function) (map[string]*callgraph.Graph, error) {
	cgs := map[string]*callgraph.Graph{}
	for _, target := range targets {
		cg := builder(ssaProg)
		cg.DeleteSyntheticNodes()

		var ok bool
		if cg.Root, ok = cg.Nodes[target]; !ok {
			return nil, fmt.Errorf("failed to find callgraph node for function %s", target.Name())
		}

		cgs[target.Name()] = cg
	}

	return cgs, nil
}

func getDependencies(cg *callgraph.Graph) ([]function, error) {
	rootModule, hasRootModule, err := getNodeModule(cg.Root)
	if err != nil {
		return nil, err
	} else if !hasRootModule {
		return nil, fmt.Errorf("root function is not in a module")
	}

	dependencies := []function{}

	visited := map[*callgraph.Node]bool{}
	queue := []*callgraph.Node{cg.Root}

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		if visited[current] {
			continue
		}

		module, hasModule, err := getNodeModule(current)
		if err != nil {
			return nil, err
		}
		if !hasModule {
			continue
		}
		if module != rootModule {
			continue
		}

		dependencies = append(dependencies, function{
			ModuleName: module,
			PkgName:    current.Func.Pkg.Pkg.Name(),
			PkgPath:    current.Func.Pkg.Pkg.Path(),
			FuncName:   current.Func.Name(),
		})

		for _, edge := range current.Out {
			if !visited[edge.Callee] {
				queue = append(queue, edge.Callee)
			}
		}

		visited[current] = true
	}

	return dependencies, nil
}

type knownPackage struct {
	hasModule bool
	module    string
}

var knownPackages = map[string]knownPackage{}

func getNodeModule(node *callgraph.Node) (string, bool, error) {
	if node.Func == nil || node.Func.Pkg == nil || node.Func.Pkg.Pkg == nil {
		return "", false, nil
	}

	pkgPath := node.Func.Pkg.Pkg.Path()
	if known, ok := knownPackages[pkgPath]; ok {
		return known.module, known.hasModule, nil
	}

	pkgs, err := packages.Load(&packages.Config{
		Mode: packages.NeedName | packages.NeedModule,
	}, pkgPath)
	if err != nil {
		return "", false, err
	}

	if len(pkgs) == 0 || pkgs[0].Module == nil {
		knownPackages[pkgPath] = knownPackage{
			hasModule: false,
			module:    "",
		}
	} else {
		knownPackages[pkgPath] = knownPackage{
			hasModule: true,
			module:    pkgs[0].Module.Path,
		}
	}

	return knownPackages[pkgPath].module, knownPackages[pkgPath].hasModule, nil
}
