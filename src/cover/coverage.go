package cover

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"golang.org/x/tools/go/ssa"
)

const mode = "set"

func calculateFunctionCoverages(path, target string, dependenciesByTarget map[functionID][]dependency) ([]Coverage, error) {
	dependencies := collapseDependencies(dependenciesByTarget)

	coverageFile, err := runTests(path, target, dependencies)
	if err != nil {
		return nil, fmt.Errorf("failed to get coverage: %v", err)
	}
	defer os.Remove(coverageFile.Name())

	coverage, err := calculateFunctionCoverageFromFile(coverageFile, dependencies)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate coverage: %v", err)
	}

	return coverage, nil
}

func collapseDependencies(dependencies map[functionID][]dependency) []dependency {
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

func runTests(path, target string, dependencies []dependency) (*os.File, error) {
	packages := make([]string, len(dependencies))
	for i, dependency := range dependencies {
		packages[i] = dependency.pkgPath
	}

	coverageFile, err := os.CreateTemp("", "deepcover-*.out")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %v", err)
	}

	cmd := exec.Command(
		"go", "test",
		"-run", target,
		"-coverprofile="+coverageFile.Name(),
		"-covermode="+mode,
		"-coverpkg="+strings.Join(packages, ","),
		path,
	)
	if err := cmd.Run(); err != nil {
		os.Remove(coverageFile.Name())
		return nil, fmt.Errorf("failed to run tests: %v", err)
	}

	return coverageFile, nil
}

func calculateFunctionCoverageFromFile(coverageFile *os.File, dependencies []dependency) ([]Coverage, error) {
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
			if strings.Contains(funcCoverage.Path, dependency.pkgPath) && funcCoverage.Name == dependency.funcName {
				if dependency.ast != nil {
					funcCoverage.Statements = countFunctionStatements(dependency.ssaFunction)
				}
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

func approxTotalCoverage(coverage []Coverage) float64 {
	if len(coverage) == 0 {
		return 0
	}

	var total float64 = 0
	var covered float64 = 0
	for _, c := range coverage {
		if c.Statements == 0 {
			continue
		}
		total += float64(c.Statements)
		covered += float64(c.Statements) * c.Coverage / 100
	}

	if total == 0 {
		return 0
	}
	return covered / total * 100
}

func countFunctionStatements(fn *ssa.Function) int {
	return len(fn.Blocks)
}
