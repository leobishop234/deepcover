package out

import (
	"strings"
	"testing"

	"github.com/leobishop234/deepcover/src/cover"
	"github.com/stretchr/testify/assert"
)

var terminalTestCoverage = cover.Result{
	Coverage: []cover.Coverage{
		{
			Path:     "example/path/file1.go",
			Name:     "package.Function1",
			Coverage: 100,
		},
		{
			Path:     "example/path/file2.go",
			Name:     "package.Function2",
			Coverage: 50,
		},
		{
			Path:     "example/path/file3.go",
			Name:     "package.Function3",
			Coverage: 0,
		},
	},
	ApproxTotalCoverage: 50,
}

func TestOutputTerminal(t *testing.T) {
	OutputTerminal(terminalTestCoverage)
}

func TestFormatTerminal(t *testing.T) {
	result := formatTerminal(terminalTestCoverage)

	assert.Contains(t, result, "PATH")
	assert.Contains(t, result, "FUNCTION")
	assert.Contains(t, result, "COVERAGE")

	assert.Contains(t, result, "package.Function1")
	assert.Contains(t, result, "package.Function2")
	assert.Contains(t, result, "package.Function3")

	assert.Contains(t, result, "100.0%")
	assert.Contains(t, result, "50.0%")
	assert.Contains(t, result, "0.0%")

	assert.Contains(t, result, "Approximate Total: 50.00%")

	lines := strings.Split(strings.TrimSpace(result), "\n")
	assert.Equal(t, 7, len(lines))
}
