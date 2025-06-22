package cover

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

type Coverage struct {
	Path     string
	Name     string
	Coverage float64
}

func getCoverage(path string, dependencies map[string][]dependency) (map[string][]Coverage, error) {
	results := make(map[string][]Coverage, len(dependencies))
	for target, dependencies := range dependencies {
		funcCoverages, err := getTestCoverage(path, target, dependencies)
		if err != nil {
			return nil, fmt.Errorf("failed to get coverage: %v", err)
		}
		results[target] = funcCoverages
	}

	return results, nil
}

func getTestCoverage(path, target string, dependencies []dependency) ([]Coverage, error) {
	output, err := runTest(path, target, dependencies)
	if err != nil {
		return nil, fmt.Errorf("failed to get coverage: %v", err)
	}

	coverage, err := parseCoverage(output, dependencies)
	if err != nil {
		return nil, fmt.Errorf("failed to parse coverage: %v", err)
	}

	return coverage, nil
}

func runTest(path, target string, dependencies []dependency) ([]byte, error) {
	packages := make(map[string]bool)
	for _, dependency := range dependencies {
		packages[dependency.PkgPath] = true
	}

	packagesList := make([]string, 0, len(packages))
	for pkg := range packages {
		packagesList = append(packagesList, pkg)
	}

	tmpDir, err := os.MkdirTemp("", "deepcover-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	coverageFile := filepath.Join(tmpDir, "coverage.out")
	cmd := exec.Command(
		"go", "test",
		"-run", target,
		"-coverprofile="+coverageFile,
		"-covermode=atomic",
		"-coverpkg="+strings.Join(packagesList, ","),
		path,
	)
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to run tests: %v", err)
	}

	output, err := exec.Command("go", "tool", "cover", "-func="+coverageFile).Output()
	if err != nil {
		return nil, fmt.Errorf("failed to parse coverage: %v", err)
	}

	return output, nil
}

func parseCoverage(output []byte, dependencies []dependency) ([]Coverage, error) {
	coverageRows := strings.Split(string(output), "\n")

	funcCoverages := []Coverage{}
	for _, row := range coverageRows {
		funcCoverage, ok, err := parseCoverageRow(row)
		if err != nil {
			return nil, fmt.Errorf("failed to extract coverage: %v", err)
		}
		if !ok {
			continue
		}

		for _, dependency := range dependencies {
			if strings.Contains(funcCoverage.Path, dependency.PkgPath) && funcCoverage.Name == dependency.FuncName {
				funcCoverages = append(funcCoverages, funcCoverage)
				break
			}
		}
	}

	return funcCoverages, nil
}

var coverageRowRegex = regexp.MustCompile(`\t+`)

func parseCoverageRow(row string) (Coverage, bool, error) {
	if row == "" || strings.HasPrefix(strings.ToLower(row), "total") {
		return Coverage{}, false, nil
	}

	row = strings.TrimSpace(row)
	row = coverageRowRegex.ReplaceAllString(row, "\t")

	parts := strings.Split(row, "\t")
	if len(parts) < 3 {
		return Coverage{}, false, nil
	}

	coverageStr := strings.TrimSuffix(parts[2], "%")
	coverage, err := strconv.ParseFloat(coverageStr, 64)
	if err != nil {
		return Coverage{}, false, fmt.Errorf("invalid coverage percentage %q: %w", parts[2], err)
	}

	return Coverage{
		Path:     parts[0],
		Name:     parts[1],
		Coverage: coverage,
	}, true, nil
}
