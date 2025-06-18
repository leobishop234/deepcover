package src

import (
	"fmt"
	"go/token"

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

func GetDependencyFunctions(path, TargetFunction string) ([]function, error) {
	cg, err := generateCallgraph(path, TargetFunction)
	if err != nil {
		return nil, err
	}

	rootModule, hasRootModule, err := getNodeModule(cg.Root)
	if err != nil {
		return nil, err
	} else if !hasRootModule {
		return nil, fmt.Errorf("root function is not in a module")
	}

	dependencies, err := getDependencies(cg, rootModule, TargetFunction)
	if err != nil {
		return nil, err
	}

	return dependencies, nil
}

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

	for _, pkg := range pkgs {
		if len(pkg.Errors) > 0 {
			return nil, fmt.Errorf("failed to load package %s: %v", pkg.PkgPath, pkg.Errors)
		}
	}

	ssaProg, ssaPkgs := ssautil.AllPackages(pkgs, 0)
	ssaProg.Build()

	var targetFunc *ssa.Function
	for _, ssaPkg := range ssaPkgs {
		for _, member := range ssaPkg.Members {
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

func getDependencies(cg *callgraph.Graph, rootModule string, targetFunction string) ([]function, error) {
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
				fmt.Println(edge.String())
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

	pkg := node.Func.Pkg.Pkg.Path()

	if known, ok := knownPackages[pkg]; ok {
		return known.module, known.hasModule, nil
	}

	pkgs, err := packages.Load(&packages.Config{
		Mode: packages.NeedName | packages.NeedModule,
	}, pkg)
	if err != nil {
		return "", false, err
	}

	if len(pkgs) == 0 || pkgs[0].Module == nil {
		knownPackages[pkg] = knownPackage{
			hasModule: false,
			module:    "",
		}
	} else {
		knownPackages[pkg] = knownPackage{
			hasModule: true,
			module:    pkgs[0].Module.Path,
		}
	}

	return knownPackages[pkg].module, knownPackages[pkg].hasModule, nil
}
