package cover

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildDataset(t *testing.T) {
	tests := []struct {
		name        string
		path        string
		regex       string
		expectFuncs []functionID
		expectError bool
	}{
		// Basic function matching tests
		{
			name:  "match Top function",
			path:  "github.com/leobishop234/deepcover/src/cover/test_data",
			regex: "Top",
			expectFuncs: []functionID{
				{pkgPath: "github.com/leobishop234/deepcover/src/cover/test_data", funcName: "Top"},
				{pkgPath: "github.com/leobishop234/deepcover/src/cover/test_data", funcName: "TestTop"},
			},
			expectError: false,
		},
		{
			name:  "match Bottom function",
			path:  "github.com/leobishop234/deepcover/src/cover/test_data",
			regex: "Bottom",
			expectFuncs: []functionID{
				{pkgPath: "github.com/leobishop234/deepcover/src/cover/test_data", funcName: "Bottom"},
				{pkgPath: "github.com/leobishop234/deepcover/src/cover/test_data", funcName: "TestBottom"},
			},
			expectError: false,
		},
		{
			name:  "match Alternative function",
			path:  "github.com/leobishop234/deepcover/src/cover/test_data",
			regex: "Alternative",
			expectFuncs: []functionID{
				{pkgPath: "github.com/leobishop234/deepcover/src/cover/test_data", funcName: "Alternative"},
				{pkgPath: "github.com/leobishop234/deepcover/src/cover/test_data", funcName: "TestAlternative"},
			},
			expectError: false,
		},
		{
			name:  "match functions starting with T",
			path:  "github.com/leobishop234/deepcover/src/cover/test_data",
			regex: "^T",
			expectFuncs: []functionID{
				{pkgPath: "github.com/leobishop234/deepcover/src/cover/test_data", funcName: "Top"},
				{pkgPath: "github.com/leobishop234/deepcover/src/cover/test_data", funcName: "TestTop"},
				{pkgPath: "github.com/leobishop234/deepcover/src/cover/test_data", funcName: "TestBottom"},
				{pkgPath: "github.com/leobishop234/deepcover/src/cover/test_data", funcName: "TestAlternative"},
			},
			expectError: false,
		},
		{
			name:  "match functions ending with e",
			path:  "github.com/leobishop234/deepcover/src/cover/test_data",
			regex: "e$",
			expectFuncs: []functionID{
				{pkgPath: "github.com/leobishop234/deepcover/src/cover/test_data", funcName: "Alternative"},
				{pkgPath: "github.com/leobishop234/deepcover/src/cover/test_data", funcName: "newInterface"},
				{pkgPath: "github.com/leobishop234/deepcover/src/cover/test_data", funcName: "TestAlternative"},
			},
			expectError: false,
		},
		{
			name:  "match all functions with wildcard",
			path:  "github.com/leobishop234/deepcover/src/cover/test_data",
			regex: ".*",
			expectFuncs: []functionID{
				{pkgPath: "github.com/leobishop234/deepcover/src/cover/test_data", funcName: "Top"},
				{pkgPath: "github.com/leobishop234/deepcover/src/cover/test_data", funcName: "Bottom"},
				{pkgPath: "github.com/leobishop234/deepcover/src/cover/test_data", funcName: "Alternative"},
				{pkgPath: "github.com/leobishop234/deepcover/src/cover/test_data", funcName: "newInterface"},
				{pkgPath: "github.com/leobishop234/deepcover/src/cover/test_data", funcName: "TestTop"},
				{pkgPath: "github.com/leobishop234/deepcover/src/cover/test_data", funcName: "TestBottom"},
				{pkgPath: "github.com/leobishop234/deepcover/src/cover/test_data", funcName: "TestAlternative"},
			},
			expectError: false,
		},
		{
			name:        "match no functions with impossible regex",
			path:        "github.com/leobishop234/deepcover/src/cover/test_data",
			regex:       "^ImpossibleFunction$",
			expectFuncs: []functionID{},
			expectError: false,
		},
		{
			name:  "match functions containing 'Test'",
			path:  "github.com/leobishop234/deepcover/src/cover/test_data",
			regex: "Test",
			expectFuncs: []functionID{
				{pkgPath: "github.com/leobishop234/deepcover/src/cover/test_data", funcName: "TestTop"},
				{pkgPath: "github.com/leobishop234/deepcover/src/cover/test_data", funcName: "TestBottom"},
				{pkgPath: "github.com/leobishop234/deepcover/src/cover/test_data", funcName: "TestAlternative"},
			},
			expectError: false,
		},
		{
			name:  "match functions with case insensitive pattern",
			path:  "github.com/leobishop234/deepcover/src/cover/test_data",
			regex: "(?i)top",
			expectFuncs: []functionID{
				{pkgPath: "github.com/leobishop234/deepcover/src/cover/test_data", funcName: "Top"},
				{pkgPath: "github.com/leobishop234/deepcover/src/cover/test_data", funcName: "TestTop"},
			},
			expectError: false,
		},

		// Subpackage and interface tests
		{
			name:        "match subpackage function",
			path:        "github.com/leobishop234/deepcover/src/cover/test_data",
			regex:       "SubPkg",
			expectFuncs: []functionID{},
			expectError: false,
		},
		{
			name:        "match interface method",
			path:        "github.com/leobishop234/deepcover/src/cover/test_data",
			regex:       "Method",
			expectFuncs: []functionID{},
			expectError: false,
		},
		{
			name:  "match constructor function",
			path:  "github.com/leobishop234/deepcover/src/cover/test_data",
			regex: "newInterface",
			expectFuncs: []functionID{
				{pkgPath: "github.com/leobishop234/deepcover/src/cover/test_data", funcName: "newInterface"},
			},
			expectError: false,
		},

		// Error handling tests
		{
			name:        "non-existent path",
			path:        "non_existent_path",
			regex:       ".*",
			expectFuncs: []functionID{},
			expectError: true,
		},
		{
			name:        "path with no Go files",
			path:        "test_data/empty_dir",
			regex:       ".*",
			expectFuncs: []functionID{},
			expectError: true,
		},
		{
			name:        "path with syntax errors",
			path:        "test_data/syntax_error",
			regex:       ".*",
			expectFuncs: []functionID{},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			regex, err := regexp.Compile(tt.regex)
			require.NoError(t, err)

			cgs, err := buildDataset(tt.path, regex)

			if !tt.expectError {
				assert.NoError(t, err)
				assert.NotNil(t, cgs.callgraph)
				assert.NotNil(t, cgs.targetNodes)
				assert.NotNil(t, cgs.asts)
			} else {
				assert.Error(t, err)
				return
			}

			// Check that we found all expected functions
			for _, wantFunc := range tt.expectFuncs {
				if assert.Contains(t, cgs.targetNodes, wantFunc, "Expected function %v not found in target nodes", wantFunc) {
					// Verify the callgraph node exists
					gotNode := cgs.targetNodes[wantFunc]
					assert.NotNil(t, gotNode, "Callgraph node should not be nil for function %v", wantFunc)
					assert.Equal(t, wantFunc.funcName, gotNode.Func.Name(), "SSA function name should match")

					// Verify the AST exists
					if astFunc, ok := cgs.asts[wantFunc]; ok {
						assert.NotNil(t, astFunc, "AST function should not be nil for function %v", wantFunc)
						assert.Equal(t, wantFunc.funcName, astFunc.Name.Name, "AST function name should match")
					}
				}
			}

			// Check that we didn't find extra functions we weren't expecting
			assert.Equal(t, len(tt.expectFuncs), len(cgs.targetNodes), "Number of found functions should match expected")
		})
	}
}

func TestInbuiltFunctionsAreFilteredOut(t *testing.T) {
	// This test verifies that init functions are filtered out as a documented limitation
	regex, err := regexp.Compile("init")
	require.NoError(t, err)

	cgs, err := buildDataset("github.com/leobishop234/deepcover/src/cover/test_data", regex)
	require.NoError(t, err)

	// Verify that no inbuilt functions are found
	for functionID := range cgs.targetNodes {
		assert.False(t, isInbuiltFunction(functionID.funcName), "Inbuilt function %s should be filtered out but was found", functionID.funcName)
	}

	t.Logf("Confirmed that %d functions were found, none of which are init functions", len(cgs.targetNodes))
}
