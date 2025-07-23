package cover

import (
	"testing"

	"go/ast"
	"go/types"

	"github.com/stretchr/testify/assert"
	"golang.org/x/tools/go/callgraph"
	"golang.org/x/tools/go/ssa"
)

func TestExtractDependencies(t *testing.T) {
	tests := []struct {
		name           string
		setupCallGraph func() callgraphDataset
		expectedDeps   []dependency
		expectedError  bool
	}{
		{
			name: "call graph with nil root",
			setupCallGraph: func() callgraphDataset {
				return callgraphDataset{
					callgraph:   &callgraph.Graph{Root: nil},
					targetNodes: make(map[functionID]*callgraph.Node),
					asts:        make(map[functionID]*ast.FuncDecl),
				}
			},
			expectedDeps:  nil,
			expectedError: true,
		},
		{
			name: "root function not in a module",
			setupCallGraph: func() callgraphDataset {
				knownPackages = map[string]knownPackage{}

				pkg := types.NewPackage("non/existent/package", "nonexistent")
				ssaPkg := &ssa.Package{Pkg: pkg}

				rootFunc := &ssa.Function{}
				rootFunc.Pkg = ssaPkg

				root := &callgraph.Node{Func: rootFunc}
				return callgraphDataset{
					callgraph:   &callgraph.Graph{Root: root},
					targetNodes: make(map[functionID]*callgraph.Node),
					asts:        make(map[functionID]*ast.FuncDecl),
				}
			},
			expectedDeps:  nil,
			expectedError: true,
		},
		{
			name: "single function in module",
			setupCallGraph: func() callgraphDataset {
				// Pre-populate cache to simulate a package that is in a module
				knownPackages = map[string]knownPackage{
					"github.com/leobishop234/deepcover/src/cover": {
						hasModule: true,
						module:    "github.com/leobishop234/deepcover",
					},
				}

				pkg := types.NewPackage("github.com/leobishop234/deepcover/src/cover", "cover")
				ssaPkg := &ssa.Package{Pkg: pkg}

				rootFunc := &ssa.Function{}
				rootFunc.Pkg = ssaPkg

				root := &callgraph.Node{Func: rootFunc}

				// Create a dummy AST function declaration
				funcDecl := &ast.FuncDecl{
					Name: &ast.Ident{Name: ""},
				}

				// Create the functionID for this function
				funcID := functionID{
					pkgPath:  "github.com/leobishop234/deepcover/src/cover",
					funcName: "", // Name() will return empty string for empty ssa.Function
				}

				return callgraphDataset{
					callgraph: &callgraph.Graph{Root: root},
					targetNodes: map[functionID]*callgraph.Node{
						funcID: root,
					},
					asts: map[functionID]*ast.FuncDecl{
						funcID: funcDecl,
					},
				}
			},
			expectedDeps: []dependency{
				{
					ModuleName: "github.com/leobishop234/deepcover",
					functionID: functionID{
						pkgPath:  "github.com/leobishop234/deepcover/src/cover",
						funcName: "", // Name() will return empty string for empty ssa.Function
					},
				},
			},
			expectedError: false,
		},
		{
			name: "multiple functions in same module",
			setupCallGraph: func() callgraphDataset {
				knownPackages = map[string]knownPackage{
					"github.com/leobishop234/deepcover/src/cover": {
						hasModule: true,
						module:    "github.com/leobishop234/deepcover",
					},
				}

				pkg := types.NewPackage("github.com/leobishop234/deepcover/src/cover", "cover")
				ssaPkg := &ssa.Package{Pkg: pkg}

				rootFunc := &ssa.Function{}
				rootFunc.Pkg = ssaPkg
				root := &callgraph.Node{Func: rootFunc}

				calledFunc := &ssa.Function{}
				calledFunc.Pkg = ssaPkg
				called := &callgraph.Node{Func: calledFunc}

				edge := &callgraph.Edge{Caller: root, Callee: called}
				root.Out = []*callgraph.Edge{edge}
				called.In = []*callgraph.Edge{edge}

				// Create dummy AST function declarations
				rootFuncDecl := &ast.FuncDecl{
					Name: &ast.Ident{Name: ""},
				}
				calledFuncDecl := &ast.FuncDecl{
					Name: &ast.Ident{Name: ""},
				}

				// Create functionIDs for both functions
				rootFuncID := functionID{
					pkgPath:  "github.com/leobishop234/deepcover/src/cover",
					funcName: "",
				}
				calledFuncID := functionID{
					pkgPath:  "github.com/leobishop234/deepcover/src/cover",
					funcName: "",
				}

				return callgraphDataset{
					callgraph: &callgraph.Graph{Root: root},
					targetNodes: map[functionID]*callgraph.Node{
						rootFuncID:   root,
						calledFuncID: called,
					},
					asts: map[functionID]*ast.FuncDecl{
						rootFuncID:   rootFuncDecl,
						calledFuncID: calledFuncDecl,
					},
				}
			},
			expectedDeps: []dependency{
				{
					ModuleName: "github.com/leobishop234/deepcover",
					functionID: functionID{
						pkgPath:  "github.com/leobishop234/deepcover/src/cover",
						funcName: "",
					},
				},
				{
					ModuleName: "github.com/leobishop234/deepcover",
					functionID: functionID{
						pkgPath:  "github.com/leobishop234/deepcover/src/cover",
						funcName: "",
					},
				},
			},
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cg := tt.setupCallGraph()

			deps, err := extractDependencies(cg, cg.callgraph.Root)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, deps)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, len(tt.expectedDeps), len(deps))

				for i, expectedDep := range tt.expectedDeps {
					if i < len(deps) {
						actualDep := deps[i]
						assert.Equal(t, expectedDep.ModuleName, actualDep.ModuleName)
						assert.Equal(t, expectedDep.pkgPath, actualDep.pkgPath)
						assert.Equal(t, expectedDep.funcName, actualDep.funcName)
						assert.NotNil(t, actualDep.node)
						assert.NotNil(t, actualDep.ast)
					}
				}
			}
		})
	}
}

func TestGetNodeModule(t *testing.T) {
	tests := []struct {
		name       string
		node       *callgraph.Node
		wantModule string
		wantHasMod bool
		wantErr    bool
	}{
		{
			name:       "nil node",
			node:       nil,
			wantModule: "",
			wantHasMod: false,
			wantErr:    false,
		},
		{
			name:       "node with nil Func",
			node:       &callgraph.Node{Func: nil},
			wantModule: "",
			wantHasMod: false,
			wantErr:    false,
		},
		{
			name: "node with nil Pkg",
			node: &callgraph.Node{
				Func: &ssa.Function{Pkg: nil},
			},
			wantModule: "",
			wantHasMod: false,
			wantErr:    false,
		},
		{
			name: "node with nil Pkg.Pkg",
			node: &callgraph.Node{
				Func: &ssa.Function{
					Pkg: &ssa.Package{Pkg: nil},
				},
			},
			wantModule: "",
			wantHasMod: false,
			wantErr:    false,
		},
		{
			name: "node with valid package path",
			node: &callgraph.Node{
				Func: &ssa.Function{
					Pkg: &ssa.Package{
						Pkg: types.NewPackage("github.com/leobishop234/deepcover/src/cover/test_data", "test_data"),
					},
				},
			},
			wantModule: "github.com/leobishop234/deepcover",
			wantHasMod: true,
			wantErr:    false,
		},
		{
			name: "node with non-existent package path",
			node: &callgraph.Node{
				Func: &ssa.Function{
					Pkg: &ssa.Package{
						Pkg: types.NewPackage("non/existent/package", "non/existent/package"),
					},
				},
			},
			wantModule: "",
			wantHasMod: false,
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			knownPackages = map[string]knownPackage{}

			gotModule, gotHasMod, err := getNodeModule(tt.node)
			assert.Equal(t, tt.wantModule, gotModule)
			assert.Equal(t, tt.wantHasMod, gotHasMod)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			// For valid packages, verify they were cached
			if tt.wantHasMod && tt.node != nil && tt.node.Func != nil && tt.node.Func.Pkg != nil && tt.node.Func.Pkg.Pkg != nil {
				cached, exists := knownPackages[tt.node.Func.Pkg.Pkg.Path()]
				assert.True(t, exists)
				assert.True(t, cached.hasModule)
				assert.Equal(t, tt.wantModule, cached.module)
			}
		})
	}
}
