package cover

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

// getTestDataPath returns the absolute path to the test_data directory
func getTestDataPath() string {
	// Get the current working directory
	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	// If we're in the src/cover directory, go up to the project root
	if filepath.Base(cwd) == "cover" && filepath.Base(filepath.Dir(cwd)) == "src" {
		return filepath.Join(filepath.Dir(filepath.Dir(cwd)), "src", "cover", "test_data")
	}

	// Otherwise, assume we're in the project root
	return filepath.Join(cwd, "src", "cover", "test_data")
}

func TestCollapseDependencies(t *testing.T) {
	tests := []struct {
		name           string
		dependencies   map[string][]dependency
		expectedResult []dependency
	}{
		{
			name:           "empty dependencies",
			dependencies:   map[string][]dependency{},
			expectedResult: []dependency{},
		},
		{
			name: "single target with single dependency",
			dependencies: map[string][]dependency{
				"target1": {
					{PkgPath: "pkg1", FuncName: "func1"},
				},
			},
			expectedResult: []dependency{
				{PkgPath: "pkg1", FuncName: "func1"},
			},
		},
		{
			name: "multiple targets with unique dependencies",
			dependencies: map[string][]dependency{
				"target1": {
					{PkgPath: "pkg1", FuncName: "func1"},
				},
				"target2": {
					{PkgPath: "pkg2", FuncName: "func2"},
				},
			},
			expectedResult: []dependency{
				{PkgPath: "pkg1", FuncName: "func1"},
				{PkgPath: "pkg2", FuncName: "func2"},
			},
		},
		{
			name: "multiple targets with overlapping dependencies",
			dependencies: map[string][]dependency{
				"target1": {
					{PkgPath: "pkg1", FuncName: "func1"},
					{PkgPath: "pkg2", FuncName: "func2"},
				},
				"target2": {
					{PkgPath: "pkg2", FuncName: "func2"},
					{PkgPath: "pkg3", FuncName: "func3"},
				},
			},
			expectedResult: []dependency{
				{PkgPath: "pkg1", FuncName: "func1"},
				{PkgPath: "pkg2", FuncName: "func2"},
				{PkgPath: "pkg3", FuncName: "func3"},
			},
		},
		{
			name: "duplicate dependencies within same target",
			dependencies: map[string][]dependency{
				"target1": {
					{PkgPath: "pkg1", FuncName: "func1"},
					{PkgPath: "pkg1", FuncName: "func1"},
				},
			},
			expectedResult: []dependency{
				{PkgPath: "pkg1", FuncName: "func1"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := collapseDependencies(tt.dependencies)

			// Sort both slices for comparison since order doesn't matter
			assert.ElementsMatch(t, tt.expectedResult, result)
		})
	}
}

func TestRunTests(t *testing.T) {
	tests := []struct {
		name                string
		path                string
		target              string
		dependencies        []dependency
		expectError         bool
		expectErrorContains string
	}{
		{
			name:                "empty dependencies",
			path:                getTestDataPath(),
			target:              "TestTop",
			dependencies:        []dependency{},
			expectError:         false,
			expectErrorContains: "",
		},
		{
			name:   "non-existent path",
			path:   "non_existent_path",
			target: "TestFunction",
			dependencies: []dependency{
				{PkgPath: "github.com/example/pkg", FuncName: "Function"},
			},
			expectError:         true,
			expectErrorContains: "failed to run tests",
		},
		{
			name:   "non-existent target",
			path:   getTestDataPath(),
			target: "NonExistentTest",
			dependencies: []dependency{
				{PkgPath: "github.com/leobishop234/deepcover/src/cover/test_data", FuncName: "Top"},
			},
			expectError:         false,
			expectErrorContains: "",
		},
		{
			name:   "successful test with Top function",
			path:   getTestDataPath(),
			target: "TestTop",
			dependencies: []dependency{
				{PkgPath: "github.com/leobishop234/deepcover/src/cover/test_data", FuncName: "Top"},
				{PkgPath: "github.com/leobishop234/deepcover/src/cover/test_data", FuncName: "Bottom"},
			},
			expectError:         false,
			expectErrorContains: "",
		},
		{
			name:   "successful test with Bottom function",
			path:   getTestDataPath(),
			target: "TestBottom",
			dependencies: []dependency{
				{PkgPath: "github.com/leobishop234/deepcover/src/cover/test_data", FuncName: "Bottom"},
				{PkgPath: "github.com/leobishop234/deepcover/src/cover/test_data/subpkg", FuncName: "SubPkg"},
			},
			expectError:         false,
			expectErrorContains: "",
		},
		{
			name:   "successful test with Alternative function",
			path:   getTestDataPath(),
			target: "TestAlternative",
			dependencies: []dependency{
				{PkgPath: "github.com/leobishop234/deepcover/src/cover/test_data", FuncName: "Alternative"},
				{PkgPath: "github.com/leobishop234/deepcover/src/cover/test_data/subpkg", FuncName: "SubPkg"},
			},
			expectError:         false,
			expectErrorContains: "",
		},
		{
			name:   "test with interface method dependency",
			path:   getTestDataPath(),
			target: "TestTop",
			dependencies: []dependency{
				{PkgPath: "github.com/leobishop234/deepcover/src/cover/test_data", FuncName: "Top"},
				{PkgPath: "github.com/leobishop234/deepcover/src/cover/test_data", FuncName: "Bottom"},
				{PkgPath: "github.com/leobishop234/deepcover/src/cover/test_data", FuncName: "newInterface"},
				{PkgPath: "github.com/leobishop234/deepcover/src/cover/test_data", FuncName: "(*Struct).Method"},
			},
			expectError:         false,
			expectErrorContains: "",
		},
		{
			name:   "test with partial dependencies",
			path:   getTestDataPath(),
			target: "TestTop",
			dependencies: []dependency{
				{PkgPath: "github.com/leobishop234/deepcover/src/cover/test_data", FuncName: "Top"},
			},
			expectError:         false,
			expectErrorContains: "",
		},
		{
			name:   "test with non-existent package in dependencies",
			path:   getTestDataPath(),
			target: "TestTop",
			dependencies: []dependency{
				{PkgPath: "github.com/leobishop234/deepcover/src/cover/test_data", FuncName: "Top"},
				{PkgPath: "github.com/non/existent/package", FuncName: "SomeFunction"},
			},
			expectError:         false,
			expectErrorContains: "",
		},
		{
			name:   "test with subpkg dependency only",
			path:   getTestDataPath(),
			target: "TestBottom",
			dependencies: []dependency{
				{PkgPath: "github.com/leobishop234/deepcover/src/cover/test_data/subpkg", FuncName: "SubPkg"},
			},
			expectError:         false,
			expectErrorContains: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			coverage, err := runTests(tt.path, tt.target, tt.dependencies)

			if tt.expectError {
				assert.Error(t, err)
				if tt.expectErrorContains != "" {
					assert.Contains(t, err.Error(), tt.expectErrorContains)
				}
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, coverage)

			// For successful cases, verify the coverage structure
			for _, cov := range coverage {
				assert.NotEmpty(t, cov.Path)
				assert.NotEmpty(t, cov.Name)
				assert.GreaterOrEqual(t, cov.Coverage, 0.0)
				assert.LessOrEqual(t, cov.Coverage, 100.0)
			}
		})
	}
}

func TestParseCoverageRow(t *testing.T) {
	tests := []struct {
		name        string
		row         string
		expectOk    bool
		expectCover Coverage
		expectError bool
	}{
		{
			name:     "valid coverage row with 100%",
			row:      "github.com/example/pkg/function.go\tFunctionName\t100.0%",
			expectOk: true,
			expectCover: Coverage{
				Path:     "github.com/example/pkg/function.go",
				Name:     "FunctionName",
				Coverage: 100.0,
			},
			expectError: false,
		},
		{
			name:     "valid coverage row with 75.5%",
			row:      "github.com/example/pkg/function.go\tFunctionName\t75.5%",
			expectOk: true,
			expectCover: Coverage{
				Path:     "github.com/example/pkg/function.go",
				Name:     "FunctionName",
				Coverage: 75.5,
			},
			expectError: false,
		},
		{
			name:     "valid coverage row with 0%",
			row:      "github.com/example/pkg/function.go\tFunctionName\t0.0%",
			expectOk: true,
			expectCover: Coverage{
				Path:     "github.com/example/pkg/function.go",
				Name:     "FunctionName",
				Coverage: 0.0,
			},
			expectError: false,
		},
		{
			name:     "valid coverage row with decimal percentage",
			row:      "github.com/example/pkg/function.go\tFunctionName\t42.7%",
			expectOk: true,
			expectCover: Coverage{
				Path:     "github.com/example/pkg/function.go",
				Name:     "FunctionName",
				Coverage: 42.7,
			},
			expectError: false,
		},
		{
			name:     "valid coverage row with multiple tabs",
			row:      "github.com/example/pkg/function.go\t\t\tFunctionName\t\t\t100.0%",
			expectOk: true,
			expectCover: Coverage{
				Path:     "github.com/example/pkg/function.go",
				Name:     "FunctionName",
				Coverage: 100.0,
			},
			expectError: false,
		},
		{
			name:     "valid coverage row with spaces around tabs",
			row:      "github.com/example/pkg/function.go\tFunctionName\t100.0%",
			expectOk: true,
			expectCover: Coverage{
				Path:     "github.com/example/pkg/function.go",
				Name:     "FunctionName",
				Coverage: 100.0,
			},
			expectError: false,
		},
		{
			name:        "empty row",
			row:         "",
			expectOk:    false,
			expectCover: Coverage{},
			expectError: false,
		},
		{
			name:        "whitespace only row",
			row:         "   \t  \t  ",
			expectOk:    false,
			expectCover: Coverage{},
			expectError: false,
		},
		{
			name:        "total row (case insensitive)",
			row:         "total:",
			expectOk:    false,
			expectCover: Coverage{},
			expectError: false,
		},
		{
			name:        "total row with coverage",
			row:         "total:\t\t\t100.0%",
			expectOk:    false,
			expectCover: Coverage{},
			expectError: false,
		},
		{
			name:        "row with insufficient parts",
			row:         "github.com/example/pkg/function.go\tFunctionName",
			expectOk:    false,
			expectCover: Coverage{},
			expectError: false,
		},
		{
			name:        "row with invalid percentage format",
			row:         "github.com/example/pkg/function.go\tFunctionName\tinvalid%",
			expectOk:    false,
			expectCover: Coverage{},
			expectError: true,
		},
		{
			name:     "row with percentage without % symbol",
			row:      "github.com/example/pkg/function.go\tFunctionName\t100.0",
			expectOk: true,
			expectCover: Coverage{
				Path:     "github.com/example/pkg/function.go",
				Name:     "FunctionName",
				Coverage: 100.0,
			},
			expectError: false,
		},
		{
			name:        "row with non-numeric percentage",
			row:         "github.com/example/pkg/function.go\tFunctionName\tabc%",
			expectOk:    false,
			expectCover: Coverage{},
			expectError: true,
		},
		{
			name:     "row with negative percentage",
			row:      "github.com/example/pkg/function.go\tFunctionName\t-50.0%",
			expectOk: true,
			expectCover: Coverage{
				Path:     "github.com/example/pkg/function.go",
				Name:     "FunctionName",
				Coverage: -50.0,
			},
			expectError: false,
		},
		{
			name:     "row with percentage over 100",
			row:      "github.com/example/pkg/function.go\tFunctionName\t150.0%",
			expectOk: true,
			expectCover: Coverage{
				Path:     "github.com/example/pkg/function.go",
				Name:     "FunctionName",
				Coverage: 150.0,
			},
			expectError: false,
		},
		{
			name:     "row with function name containing special characters",
			row:      "github.com/example/pkg/function.go\tFunction_Name_With_Underscores\t85.3%",
			expectOk: true,
			expectCover: Coverage{
				Path:     "github.com/example/pkg/function.go",
				Name:     "Function_Name_With_Underscores",
				Coverage: 85.3,
			},
			expectError: false,
		},
		{
			name:     "row with function name containing dots",
			row:      "github.com/example/pkg/function.go\tFunction.Name.With.Dots\t67.8%",
			expectOk: true,
			expectCover: Coverage{
				Path:     "github.com/example/pkg/function.go",
				Name:     "Function.Name.With.Dots",
				Coverage: 67.8,
			},
			expectError: false,
		},
		{
			name:     "row with path containing special characters",
			row:      "github.com/example/pkg/sub-pkg/function_test.go\tTestFunction\t92.1%",
			expectOk: true,
			expectCover: Coverage{
				Path:     "github.com/example/pkg/sub-pkg/function_test.go",
				Name:     "TestFunction",
				Coverage: 92.1,
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			coverage, ok, err := parseCoverageRow(tt.row)

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.expectOk, ok)

			if tt.expectOk {
				assert.Equal(t, tt.expectCover.Path, coverage.Path)
				assert.Equal(t, tt.expectCover.Name, coverage.Name)
				assert.Equal(t, tt.expectCover.Coverage, coverage.Coverage)
			} else {
				assert.Equal(t, Coverage{}, coverage)
			}
		})
	}
}
