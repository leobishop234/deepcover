package cover

import (
	"fmt"

	"golang.org/x/tools/go/callgraph"
	"golang.org/x/tools/go/packages"
)

func getDependencies(cgs analysis) (map[functionID][]dependency, error) {
	dependencies := make(map[functionID][]dependency, len(cgs.targetNodes))
	var err error
	for targetID, targetNode := range cgs.targetNodes {
		dependencies[targetID], err = extractDependencies(cgs, targetNode)
		if err != nil {
			return nil, err
		}
	}

	return dependencies, nil
}

func extractDependencies(cg analysis, start *callgraph.Node) ([]dependency, error) {
	if start == nil {
		return nil, fmt.Errorf("start node is nil")
	}

	rootModule, hasRootModule, err := getNodeModule(start)
	if err != nil {
		return nil, err
	} else if !hasRootModule {
		return nil, fmt.Errorf("root function is not in a module")
	}

	dependencies := []dependency{}

	visited := map[*callgraph.Node]bool{}
	queue := []*callgraph.Node{start}

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
			functionID: functionID{
				pkgPath:  current.Func.Pkg.Pkg.Path(),
				funcName: current.Func.Name(),
			},
			node: current,
			ast: cg.asts[functionID{
				pkgPath:  current.Func.Pkg.Pkg.Path(),
				funcName: current.Func.Name(),
			}],
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
	if node == nil || node.Func == nil || node.Func.Pkg == nil || node.Func.Pkg.Pkg == nil {
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
