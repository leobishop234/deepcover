package src

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

type FunctionCoverage struct {
	Path     string
	Name     string
	Coverage float64
}

func GetCoverage(path, targetRxp string, expectedPackages []string) ([]FunctionCoverage, error) {
	output, err := getCoverage(path, targetRxp, expectedPackages)
	if err != nil {
		return nil, fmt.Errorf("failed to get coverage: %v", err)
	}

	coverage, err := parseCoverage(output)
	if err != nil {
		return nil, fmt.Errorf("failed to parse coverage: %v", err)
	}

	return coverage, nil
}

func getCoverage(path, targetRxp string, expectedPackages []string) ([]byte, error) {
	tmpDir, err := os.MkdirTemp("", "deepcover-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	coverageFile := filepath.Join(tmpDir, "coverage.out")
	cmd := exec.Command(
		"go", "test",
		"-run", targetRxp,
		"-coverprofile="+coverageFile,
		"-covermode=atomic",
		"-coverpkg="+strings.Join(expectedPackages, ","),
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

func parseCoverage(output []byte) ([]FunctionCoverage, error) {
	coverageRows := strings.Split(string(output), "\n")

	funcCoverages := []FunctionCoverage{}
	for _, row := range coverageRows {
		funcCoverage, ok, err := parseCoverageRow(row)
		if err != nil {
			return nil, fmt.Errorf("failed to extract coverage: %v", err)
		}
		if !ok {
			continue
		}

		funcCoverages = append(funcCoverages, funcCoverage)
	}

	return funcCoverages, nil
}

var coverageRowRegex = regexp.MustCompile(`\t+`)

func parseCoverageRow(row string) (FunctionCoverage, bool, error) {
	if row == "" || strings.HasPrefix(strings.ToLower(row), "total") {
		return FunctionCoverage{}, false, nil
	}

	row = strings.TrimSpace(row)
	row = coverageRowRegex.ReplaceAllString(row, "\t")

	parts := strings.Split(row, "\t")
	if len(parts) < 3 {
		return FunctionCoverage{}, false, nil
	}

	coverageStr := strings.TrimSuffix(parts[2], "%")
	coverage, err := strconv.ParseFloat(coverageStr, 64)
	if err != nil {
		return FunctionCoverage{}, false, fmt.Errorf("invalid coverage percentage %q: %w", parts[2], err)
	}

	return FunctionCoverage{
		Path:     parts[0],
		Name:     parts[1],
		Coverage: coverage,
	}, true, nil
}
