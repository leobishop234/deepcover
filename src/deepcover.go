package src

import (
	"fmt"
	"regexp"
)

func Deepcover(pkgPath string, targetRegex *regexp.Regexp) (map[string][]FunctionCoverage, error) {
	dependencies, err := GetDependencies(pkgPath, targetRegex)
	if err != nil {
		return nil, fmt.Errorf("failed to get dependencies: %v", err)
	}

	results := make(map[string][]FunctionCoverage, len(dependencies))
	for targetFunc, dependencies := range dependencies {
		funcCoverages, err := GetCoverage(pkgPath, targetFunc, dependencies)
		if err != nil {
			return nil, fmt.Errorf("failed to get coverage: %v", err)
		}
		results[targetFunc] = funcCoverages
	}

	return results, nil
}
