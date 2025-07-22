package cover

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestGetCoverage(t *testing.T) {
	tests := []struct {
		name                 string
		path                 string
		target               string
		dependenciesByTarget map[functionID][]dependency
		expectError          bool
		expectedCoverage     int
	}{
		{
			name:                 "empty dependencies",
			path:                 getTestDataPath(),
			target:               "TestTop",
			dependenciesByTarget: map[functionID][]dependency{},
			expectError:          false,
			expectedCoverage:     0,
		},
		{
			name:   "non-existent path",
			path:   "non_existent_path",
			target: "TestFunction",
			dependenciesByTarget: map[functionID][]dependency{
				{pkgPath: "github.com/example/pkg", funcName: "target1"}: {
					{PkgPath: "github.com/example/pkg", FuncName: "Function"},
				},
			},
			expectError:      true,
			expectedCoverage: 0,
		},
		{
			name:   "successful coverage with single target",
			path:   getTestDataPath(),
			target: "TestTop",
			dependenciesByTarget: map[functionID][]dependency{
				{pkgPath: "github.com/leobishop234/deepcover/src/cover/test_data", funcName: "target1"}: {
					{PkgPath: "github.com/leobishop234/deepcover/src/cover/test_data", FuncName: "Top"},
				},
			},
			expectError:      false,
			expectedCoverage: 1,
		},
		{
			name:   "successful coverage with multiple targets and overlapping dependencies",
			path:   getTestDataPath(),
			target: "TestTop",
			dependenciesByTarget: map[functionID][]dependency{
				{pkgPath: "github.com/leobishop234/deepcover/src/cover/test_data", funcName: "target1"}: {
					{PkgPath: "github.com/leobishop234/deepcover/src/cover/test_data", FuncName: "Top"},
					{PkgPath: "github.com/leobishop234/deepcover/src/cover/test_data", FuncName: "Bottom"},
				},
				{pkgPath: "github.com/leobishop234/deepcover/src/cover/test_data", funcName: "target2"}: {
					{PkgPath: "github.com/leobishop234/deepcover/src/cover/test_data", FuncName: "Bottom"},
				},
			},
			expectError:      false,
			expectedCoverage: 2,
		},
		{
			name:   "test with subpackage dependencies",
			path:   getTestDataPath(),
			target: "TestBottom",
			dependenciesByTarget: map[functionID][]dependency{
				{pkgPath: "github.com/leobishop234/deepcover/src/cover/test_data", funcName: "target1"}: {
					{PkgPath: "github.com/leobishop234/deepcover/src/cover/test_data", FuncName: "Bottom"},
					{PkgPath: "github.com/leobishop234/deepcover/src/cover/test_data/subpkg", FuncName: "SubPkg"},
				},
			},
			expectError:      false,
			expectedCoverage: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			coverage, err := getCoverage(tt.path, tt.target, tt.dependenciesByTarget)

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, len(coverage), tt.expectedCoverage)
		})
	}
}

func TestCollapseDependencies(t *testing.T) {
	tests := []struct {
		name           string
		dependencies   map[functionID][]dependency
		expectedResult []dependency
	}{
		{
			name:           "empty dependencies",
			dependencies:   map[functionID][]dependency{},
			expectedResult: []dependency{},
		},
		{
			name: "single target with single dependency",
			dependencies: map[functionID][]dependency{
				{pkgPath: "pkg1", funcName: "target1"}: {
					{PkgPath: "pkg1", FuncName: "func1"},
				},
			},
			expectedResult: []dependency{
				{PkgPath: "pkg1", FuncName: "func1"},
			},
		},
		{
			name: "multiple targets with unique dependencies",
			dependencies: map[functionID][]dependency{
				{pkgPath: "pkg1", funcName: "target1"}: {
					{PkgPath: "pkg1", FuncName: "func1"},
				},
				{pkgPath: "pkg2", funcName: "target2"}: {
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
			dependencies: map[functionID][]dependency{
				{pkgPath: "pkg1", funcName: "target1"}: {
					{PkgPath: "pkg1", FuncName: "func1"},
					{PkgPath: "pkg2", FuncName: "func2"},
				},
				{pkgPath: "pkg2", funcName: "target2"}: {
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
			dependencies: map[functionID][]dependency{
				{pkgPath: "pkg1", funcName: "target1"}: {
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
			assert.ElementsMatch(t, tt.expectedResult, result)
		})
	}
}

func TestRunTests(t *testing.T) {
	tests := []struct {
		name         string
		path         string
		target       string
		dependencies []dependency
		expectError  bool
	}{
		{
			name:         "empty dependencies",
			path:         getTestDataPath(),
			target:       "TestTop",
			dependencies: []dependency{},
			expectError:  false,
		},
		{
			name:   "non-existent path",
			path:   "non_existent_path",
			target: "TestFunction",
			dependencies: []dependency{
				{PkgPath: "github.com/example/pkg", FuncName: "Function"},
			},
			expectError: true,
		},
		{
			name:   "successful test with valid path and target",
			path:   getTestDataPath(),
			target: "TestTop",
			dependencies: []dependency{
				{PkgPath: "github.com/leobishop234/deepcover/src/cover/test_data", FuncName: "Top"},
			},
			expectError: false,
		},
		{
			name:   "test with multiple dependencies",
			path:   getTestDataPath(),
			target: "TestBottom",
			dependencies: []dependency{
				{PkgPath: "github.com/leobishop234/deepcover/src/cover/test_data", FuncName: "Bottom"},
				{PkgPath: "github.com/leobishop234/deepcover/src/cover/test_data/subpkg", FuncName: "SubPkg"},
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			coverageFile, err := runTests(tt.path, tt.target, tt.dependencies)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, coverageFile)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, coverageFile)

			// Verify the file exists and can be read
			_, err = os.Stat(coverageFile.Name())
			assert.NoError(t, err)

			// Clean up
			coverageFile.Close()
			os.Remove(coverageFile.Name())
		})
	}
}

func TestCalculateCoverageFromFile(t *testing.T) {
	// Create a temporary coverage file with known content
	coverageContent := `mode: atomic
github.com/leobishop234/deepcover/src/cover/test_data/example.go:5.13,7.2 1 1
github.com/leobishop234/deepcover/src/cover/test_data/example.go:9.16,11.2 1 1
github.com/leobishop234/deepcover/src/cover/test_data/subpkg/subtest.go:12.15,14.2 1 1
github.com/leobishop234/deepcover/src/cover/test_data/subpkg/subtest.go:14.2,16.2 1 0
`

	coverageFile, err := os.CreateTemp("", "test-coverage-*.out")
	require.NoError(t, err)
	defer os.Remove(coverageFile.Name())
	defer coverageFile.Close()

	_, err = coverageFile.WriteString(coverageContent)
	require.NoError(t, err)
	coverageFile.Close()

	// Reopen for reading
	coverageFile, err = os.Open(coverageFile.Name())
	require.NoError(t, err)

	tests := []struct {
		name                string
		dependencies        []dependency
		expectError         bool
		expectCoverageCount int
	}{
		{
			name:                "empty dependencies",
			dependencies:        []dependency{},
			expectError:         false,
			expectCoverageCount: 0,
		},
		{
			name: "single matching dependency",
			dependencies: []dependency{
				{PkgPath: "github.com/leobishop234/deepcover/src/cover/test_data", FuncName: "Top"},
			},
			expectError:         false,
			expectCoverageCount: 1,
		},
		{
			name: "multiple dependencies with matches",
			dependencies: []dependency{
				{PkgPath: "github.com/leobishop234/deepcover/src/cover/test_data", FuncName: "Top"},
				{PkgPath: "github.com/leobishop234/deepcover/src/cover/test_data", FuncName: "Bottom"},
				{PkgPath: "github.com/leobishop234/deepcover/src/cover/test_data/subpkg", FuncName: "SubPkg"},
			},
			expectError:         false,
			expectCoverageCount: 3,
		},
		{
			name: "dependencies with no matches",
			dependencies: []dependency{
				{PkgPath: "github.com/non/existent", FuncName: "Function"},
			},
			expectError:         false,
			expectCoverageCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			coverage, err := calculateCoverageFromFile(coverageFile, tt.dependencies)

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Len(t, coverage, tt.expectCoverageCount)
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
