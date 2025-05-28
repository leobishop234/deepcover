package src

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

type FuncCoverage struct {
	Path     string
	Name     string
	Coverage float64
}

func GetCoverage(path, targetRxp string, expectedPackages []string) ([]FuncCoverage, error) {
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

func parseCoverage(output []byte) ([]FuncCoverage, error) {
	coverageRows := strings.Split(string(output), "\n")

	funcCoverages := []FuncCoverage{}
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

func parseCoverageRow(row string) (FuncCoverage, bool, error) {
	if row == "" || strings.HasPrefix(strings.ToLower(row), "total") {
		return FuncCoverage{}, false, nil
	}

	parts := strings.Split(row, "\t")
	if len(parts) < 6 {
		return FuncCoverage{}, false, nil
	}

	path := parts[0]

	name := ""
	for i := 1; i < len(parts)-1; i++ {
		if parts[i] != "" {
			name = parts[i]
			break
		}
	}

	coverageStr := strings.TrimSuffix(parts[len(parts)-1], "%")
	coverage, err := strconv.ParseFloat(coverageStr, 64)
	if err != nil {
		return FuncCoverage{}, false, fmt.Errorf("invalid coverage percentage %q: %w", parts[len(parts)-1], err)
	}

	return FuncCoverage{
		Path:     path,
		Name:     name,
		Coverage: coverage,
	}, true, nil
}
