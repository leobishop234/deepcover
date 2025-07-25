package cover

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildCallgraphs(t *testing.T) {
	tests := []struct {
		name        string
		path        string
		regex       string
		expectFuncs []string
		expectError bool
	}{
		// Basic function matching tests
		{
			name:        "match Top function",
			path:        "github.com/leobishop234/deepcover/src/cover/test_data",
			regex:       "Top",
			expectFuncs: []string{"Top", "TestTop"},
			expectError: false,
		},
		{
			name:        "match Bottom function",
			path:        "github.com/leobishop234/deepcover/src/cover/test_data",
			regex:       "Bottom",
			expectFuncs: []string{"Bottom", "TestBottom"},
			expectError: false,
		},
		{
			name:        "match Alternative function",
			path:        "github.com/leobishop234/deepcover/src/cover/test_data",
			regex:       "Alternative",
			expectFuncs: []string{"Alternative", "TestAlternative"},
			expectError: false,
		},
		{
			name:        "match functions starting with T",
			path:        "github.com/leobishop234/deepcover/src/cover/test_data",
			regex:       "^T",
			expectFuncs: []string{"Top", "TestTop", "TestBottom", "TestAlternative"},
			expectError: false,
		},
		{
			name:        "match functions ending with e",
			path:        "github.com/leobishop234/deepcover/src/cover/test_data",
			regex:       "e$",
			expectFuncs: []string{"Alternative", "newInterface", "TestAlternative"},
			expectError: false,
		},
		{
			name:        "match all functions with wildcard",
			path:        "github.com/leobishop234/deepcover/src/cover/test_data",
			regex:       ".*",
			expectFuncs: []string{"Top", "Bottom", "Alternative", "newInterface", "TestTop", "TestBottom", "TestAlternative", "init", "init#1", "main"},
			expectError: false,
		},
		{
			name:        "match no functions with impossible regex",
			path:        "github.com/leobishop234/deepcover/src/cover/test_data",
			regex:       "^ImpossibleFunction$",
			expectFuncs: []string{},
			expectError: false,
		},
		{
			name:        "match functions containing 'Test'",
			path:        "github.com/leobishop234/deepcover/src/cover/test_data",
			regex:       "Test",
			expectFuncs: []string{"TestTop", "TestBottom", "TestAlternative"},
			expectError: false,
		},
		{
			name:        "match functions with case insensitive pattern",
			path:        "github.com/leobishop234/deepcover/src/cover/test_data",
			regex:       "(?i)top",
			expectFuncs: []string{"Top", "TestTop"},
			expectError: false,
		},

		// Subpackage and interface tests
		{
			name:        "match subpackage function",
			path:        "github.com/leobishop234/deepcover/src/cover/test_data",
			regex:       "SubPkg",
			expectFuncs: []string{},
			expectError: false,
		},
		{
			name:        "match interface method",
			path:        "github.com/leobishop234/deepcover/src/cover/test_data",
			regex:       "Method",
			expectFuncs: []string{},
			expectError: false,
		},
		{
			name:        "match constructor function",
			path:        "github.com/leobishop234/deepcover/src/cover/test_data",
			regex:       "newInterface",
			expectFuncs: []string{"newInterface"},
			expectError: false,
		},

		// Error handling tests
		{
			name:        "non-existent path",
			path:        "non_existent_path",
			regex:       ".*",
			expectFuncs: []string{},
			expectError: true,
		},
		{
			name:        "path with no Go files",
			path:        "test_data/empty_dir",
			regex:       ".*",
			expectFuncs: []string{},
			expectError: true,
		},
		{
			name:        "path with syntax errors",
			path:        "test_data/syntax_error",
			regex:       ".*",
			expectFuncs: []string{},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			regex, err := regexp.Compile(tt.regex)
			require.NoError(t, err)

			cgs, err := buildCallgraphs(tt.path, regex)

			if !tt.expectError {
				assert.NoError(t, err)
				assert.NotNil(t, cgs)
			} else {
				assert.Error(t, err)
				return
			}

			// Verify expected functions are found
			for _, wantFunc := range tt.expectFuncs {
				assert.Contains(t, cgs, wantFunc)
			}

			// Verify no unexpected functions are present
			for funcName := range cgs {
				assert.Contains(t, tt.expectFuncs, funcName)
			}

			// Verify callgraph structure for each expected function
			for _, wantFunc := range tt.expectFuncs {
				if cg, exists := cgs[wantFunc]; exists {
					if assert.NotNil(t, cg) {
						if assert.NotNil(t, cg.Root) {
							if assert.NotNil(t, cg.Root.Func) {
								assert.Equal(t, cg.Root.Func.Name(), wantFunc)
							}
						}
						assert.NotEmpty(t, cg.Nodes)
					}
				}
			}
		})
	}
}
