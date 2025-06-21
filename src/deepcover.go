package src

import (
	"regexp"
)

func Deepcover(pkgPath string, targetRegex *regexp.Regexp) (map[string][]Coverage, error) {
	cgs, err := buildCallgraphs(pkgPath, targetRegex)
	if err != nil {
		return nil, err
	}

	dependencies, err := getDependencies(cgs)
	if err != nil {
		return nil, err
	}

	return getCoverage(pkgPath, dependencies)
}
