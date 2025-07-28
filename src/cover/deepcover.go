package cover

import (
	"regexp"
)

type Result struct {
	Coverage            []Coverage
	ApproxTotalCoverage float64
}

type Coverage struct {
	Path       string
	Name       string
	Statements int
	Lines      int
	Coverage   float64
}

func Deepcover(pkgPath, target string) (Result, error) {
	targetRegex, err := regexp.Compile(target)
	if err != nil {
		return Result{}, err
	}

	cgs, err := buildAnalysis(pkgPath, targetRegex)
	if err != nil {
		return Result{}, err
	}

	dependencies, err := getDependencies(cgs)
	if err != nil {
		return Result{}, err
	}

	coverage, err := calculateFunctionCoverages(pkgPath, target, dependencies)
	if err != nil {
		return Result{}, err
	}

	return Result{
		Coverage:            coverage,
		ApproxTotalCoverage: approxTotalCoverage(coverage),
	}, nil
}
