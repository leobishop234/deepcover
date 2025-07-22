package cover

import (
	"go/ast"

	"golang.org/x/tools/go/callgraph"
)

type functionID struct {
	pkgPath  string
	funcName string
}

type callgraphDataset struct {
	callgraph   *callgraph.Graph
	targetNodes map[functionID]*callgraph.Node
	asts        map[functionID]*ast.FuncDecl
}

type dependency struct {
	ModuleName string
	PkgName    string
	PkgPath    string
	FuncName   string
}
