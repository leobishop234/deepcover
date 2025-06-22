package main

import (
	"flag"
	"fmt"
	"math"
	"os"
	"regexp"
	"strings"

	"deepcover/src/cover"
)

func main() {
	flag.Parse()
	args := flag.Args()
	if len(args) != 2 {
		fmt.Fprintf(os.Stderr, "Error: Expected 2 arguments (pkg path and target regex)\n")
		os.Exit(1)
	}

	pkgPath := args[0]
	targetRegexStr := args[1]

	targetRegex, err := regexp.Compile(targetRegexStr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Invalid target regex: %v\n", err)
		os.Exit(1)
	}

	if err := run(pkgPath, targetRegex); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run(pkgPath string, targetRegex *regexp.Regexp) error {
	if pkgPath == "" {
		return fmt.Errorf("pkg path is required")
	}

	funcCoverages, err := cover.Deepcover(pkgPath, targetRegex)
	if err != nil {
		return fmt.Errorf("failed to get dependencies: %v", err)
	}

	for target, funcCoverages := range funcCoverages {
		displayCoverage(target, funcCoverages)
	}

	return nil
}

func displayCoverage(target string, funcCoverages []cover.Coverage) {
	targetLen := int(math.Max(float64(len(target)+2), float64(len("TARGET"))))

	var pathLen, nameLen, coverageLen int
	for _, funcCoverage := range funcCoverages {
		if len(funcCoverage.Path) > pathLen {
			pathLen = len(funcCoverage.Path)
		}
		if len(funcCoverage.Name) > nameLen {
			nameLen = len(funcCoverage.Name)
		}
		if len(fmt.Sprintf("%.1f%%", funcCoverage.Coverage)) > coverageLen {
			coverageLen = len(fmt.Sprintf("%.1f%%", funcCoverage.Coverage))
		}
	}
	pathLen += 2
	nameLen += 2
	coverageLen += 2

	title := fmt.Sprintf("%-*s %-*s %-*s %-*s", targetLen, "TARGET", pathLen, "PATH", nameLen, "FUNCTION", coverageLen, "COVERAGE")
	fmt.Println(title)
	fmt.Println(strings.Repeat("-", len(title)))

	for _, funcCoverage := range funcCoverages {
		coverageStr := fmt.Sprintf("%.1f%%", funcCoverage.Coverage)
		fmt.Printf("%-*s %-*s %-*s %-*s\n",
			targetLen,
			truncateString(target, targetLen),
			pathLen,
			truncateString(funcCoverage.Path, pathLen),
			nameLen,
			truncateString(funcCoverage.Name, nameLen),
			coverageLen,
			coverageStr)
	}
	fmt.Println()
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
