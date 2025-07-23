package cover

import (
	"regexp"
)

type Coverage struct {
	Path     string
	Name     string
	Coverage float64
}

func Deepcover(pkgPath, target string) ([]Coverage, error) {
	targetRegex, err := regexp.Compile(target)
	if err != nil {
		return nil, err
	}

	cgs, err := buildAnalysis(pkgPath, targetRegex)
	if err != nil {
		return nil, err
	}

	dependencies, err := getDependencies(cgs)
	if err != nil {
		return nil, err
	}

	coverage, err := getCoverage(pkgPath, target, dependencies)
	if err != nil {
		return nil, err
	}

	return coverage, nil
}
