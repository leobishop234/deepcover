package out

import (
	"fmt"
	"strings"

	"github.com/leobishop234/deepcover/src/cover"
)

func OutputTerminal(coverage cover.Result) {
	fmt.Print(formatTerminal(coverage))
}

func formatTerminal(coverage cover.Result) string {
	var pathLen, nameLen, coverageLen int
	for _, funcCoverage := range coverage.Coverage {
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

	var result strings.Builder

	title := fmt.Sprintf("%-*s %-*s %-*s", pathLen, "PATH", nameLen, "FUNCTION", coverageLen, "COVERAGE")
	result.WriteString(title)
	result.WriteString("\n")
	result.WriteString(strings.Repeat("-", len(title)))
	result.WriteString("\n")

	for _, funcCoverage := range coverage.Coverage {
		coverageStr := fmt.Sprintf("%.1f%%", funcCoverage.Coverage)
		line := fmt.Sprintf("%-*s %-*s %-*s\n",
			pathLen,
			funcCoverage.Path,
			nameLen,
			funcCoverage.Name,
			coverageLen,
			coverageStr)
		result.WriteString(line)
	}

	result.WriteString(fmt.Sprintf("Total: %.2f%%", coverage.ApproxTotalCoverage))

	return result.String()
}
