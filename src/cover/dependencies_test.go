package cover

import (
	"testing"

	"go/types"

	"github.com/stretchr/testify/assert"
	"golang.org/x/tools/go/callgraph"
	"golang.org/x/tools/go/ssa"
)

func TestExtractDependencies(t *testing.T) {
	tests := []struct {
		name           string
		setupCallGraph func() *callgraph.Graph
		expectedDeps   []dependency
		expectedError  bool
	}{
		{
			name: "nil call graph",
			setupCallGraph: func() *callgraph.Graph {
				return nil
			},
			expectedDeps:  nil,
			expectedError: true,
		},
		{
			name: "call graph with nil root",
			setupCallGraph: func() *callgraph.Graph {
				return &callgraph.Graph{Root: nil}
			},
			expectedDeps:  nil,
			expectedError: true,
		},
		{
			name: "root function not in a module",
			setupCallGraph: func() *callgraph.Graph {
				knownPackages = map[string]knownPackage{}

				pkg := types.NewPackage("non/existent/package", "nonexistent")
				ssaPkg := &ssa.Package{Pkg: pkg}

				rootFunc := &ssa.Function{}
				rootFunc.Pkg = ssaPkg

				root := &callgraph.Node{Func: rootFunc}
				return &callgraph.Graph{Root: root}
			},
			expectedDeps:  nil,
			expectedError: true,
		},
		{
			name: "single function in module",
			setupCallGraph: func() *callgraph.Graph {
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
				return &callgraph.Graph{Root: root}
			},
			expectedDeps: []dependency{
				{
					ModuleName: "github.com/leobishop234/deepcover",
					PkgName:    "cover",
					PkgPath:    "github.com/leobishop234/deepcover/src/cover",
					FuncName:   "", // Name() will return empty string for empty ssa.Function
				},
			},
			expectedError: false,
		},
		{
			name: "multiple functions in same module",
			setupCallGraph: func() *callgraph.Graph {
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

				return &callgraph.Graph{Root: root}
			},
			expectedDeps: []dependency{
				{
					ModuleName: "github.com/leobishop234/deepcover",
					PkgName:    "cover",
					PkgPath:    "github.com/leobishop234/deepcover/src/cover",
					FuncName:   "",
				},
				{
					ModuleName: "github.com/leobishop234/deepcover",
					PkgName:    "cover",
					PkgPath:    "github.com/leobishop234/deepcover/src/cover",
					FuncName:   "",
				},
			},
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cg := tt.setupCallGraph()

			deps, err := extractDependencies(cg)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, deps)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, len(tt.expectedDeps), len(deps))

				expectedMap := make(map[string]dependency)
				for _, dep := range tt.expectedDeps {
					key := dep.PkgPath + ":" + dep.FuncName
					expectedMap[key] = dep
				}

				actualMap := make(map[string]dependency)
				for _, dep := range deps {
					key := dep.PkgPath + ":" + dep.FuncName
					actualMap[key] = dep
				}

				assert.Equal(t, expectedMap, actualMap)
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
