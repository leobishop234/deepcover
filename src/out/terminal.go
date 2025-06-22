package out

import (
	"deepcover/src/cover"
	"fmt"
	"strings"
)

func PrintCoverage(coverage []cover.Coverage) {
	var pathLen, nameLen, coverageLen int
	for _, funcCoverage := range coverage {
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

	title := fmt.Sprintf("%-*s %-*s %-*s", pathLen, "PATH", nameLen, "FUNCTION", coverageLen, "COVERAGE")
	fmt.Println(title)
	fmt.Println(strings.Repeat("-", len(title)))

	for _, funcCoverage := range coverage {
		coverageStr := fmt.Sprintf("%.1f%%", funcCoverage.Coverage)
		fmt.Printf("%-*s %-*s %-*s\n",
			pathLen,
			funcCoverage.Path,
			nameLen,
			funcCoverage.Name,
			coverageLen,
			coverageStr)
	}
	fmt.Println()
}
