package cover

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

type Coverage struct {
	Path     string
	Name     string
	Coverage float64
}

func getCoverage(path, target string, dependenciesByTarget map[string][]dependency) ([]Coverage, error) {
	dependencies := collapseDependencies(dependenciesByTarget)
	coverage, err := runTests(path, target, dependencies)
	if err != nil {
		return nil, fmt.Errorf("failed to get coverage: %v", err)
	}

	return coverage, nil
}

func collapseDependencies(dependencies map[string][]dependency) []dependency {
	depMap := make(map[dependency]bool)
	for _, deps := range dependencies {
		for _, dep := range deps {
			depMap[dep] = true
		}
	}

	collapsed := make([]dependency, 0, len(depMap))
	for dep := range depMap {
		collapsed = append(collapsed, dep)
	}

	return collapsed
}

func runTests(path, target string, dependencies []dependency) ([]Coverage, error) {
	packages := make([]string, len(dependencies))
	for i, dependency := range dependencies {
		packages[i] = dependency.PkgPath
	}

	coverageFile, err := os.CreateTemp("", "deepcover-*.out")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %v", err)
	}
	defer os.Remove(coverageFile.Name())

	cmd := exec.Command(
		"go", "test",
		"-run", target,
		"-coverprofile="+coverageFile.Name(),
		"-covermode=atomic",
		"-coverpkg="+strings.Join(packages, ","),
		path,
	)
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to run tests: %v", err)
	}

	output, err := exec.Command(
		"go", "tool", "cover",
		"-func="+coverageFile.Name(),
	).Output()
	if err != nil {
		return nil, fmt.Errorf("failed to parse coverage: %v", err)
	}

	rows := strings.Split(string(output), "\n")
	coverage := []Coverage{}
	for _, row := range rows {
		funcCoverage, ok, err := parseCoverageRow(row)
		if err != nil {
			return nil, fmt.Errorf("failed to extract coverage: %v", err)
		}
		if !ok {
			continue
		}

		for _, dependency := range dependencies {
			if strings.Contains(funcCoverage.Path, dependency.PkgPath) && funcCoverage.Name == dependency.FuncName {
				coverage = append(coverage, funcCoverage)
				break
			}
		}
	}

	return coverage, nil
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
