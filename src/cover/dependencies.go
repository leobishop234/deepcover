package cover

import (
	"fmt"

	"golang.org/x/tools/go/callgraph"
	"golang.org/x/tools/go/packages"
)

type dependency struct {
	ModuleName string
	PkgName    string
	PkgPath    string
	FuncName   string
}

func getDependencies(cgs map[string]*callgraph.Graph) (map[string][]dependency, error) {
	dependencies := make(map[string][]dependency, len(cgs))
	var err error
	for target, cg := range cgs {
		dependencies[target], err = extractDependencies(cg)
		if err != nil {
			return nil, err
		}
	}

	return dependencies, nil
}

func extractDependencies(cg *callgraph.Graph) ([]dependency, error) {
	rootModule, hasRootModule, err := getNodeModule(cg.Root)
	if err != nil {
		return nil, err
	} else if !hasRootModule {
		return nil, fmt.Errorf("root function is not in a module")
	}

	dependencies := []dependency{}

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

		dependencies = append(dependencies, dependency{
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
