package cover

import (
	"go/ast"

	"golang.org/x/tools/go/callgraph"
	"golang.org/x/tools/go/ssa"
)

type functionID struct {
	pkgPath  string
	funcName string
}

type analysis struct {
	callgraph   *callgraph.Graph
	targetNodes map[functionID]*callgraph.Node
	asts        map[functionID]*ast.FuncDecl
}

type dependency struct {
	ModuleName string
	functionID
	ssaFunction *ssa.Function
	node        *callgraph.Node
	ast         *ast.FuncDecl
}
