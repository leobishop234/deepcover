package out

import (
	"strings"
	"testing"

	"github.com/leobishop234/deepcover/src/cover"
	"github.com/stretchr/testify/assert"
)

var terminalTestCoverage = []cover.Coverage{
	{
		Path:     "test_data/example.go",
		Name:     "test_data.Top",
		Coverage: 100,
	},
	{
		Path:     "test_data/example.go",
		Name:     "test_data.Bottom",
		Coverage: 50,
	},
	{
		Path:     "test_data/example.go",
		Name:     "test_data.Alternative",
		Coverage: 0,
	},
}

func TestOutputTerminal(t *testing.T) {
	OutputTerminal(terminalTestCoverage)
}

func TestFormatTerminal(t *testing.T) {
	result := formatTerminal(terminalTestCoverage)

	assert.Contains(t, result, "PATH")
	assert.Contains(t, result, "FUNCTION")
	assert.Contains(t, result, "COVERAGE")

	assert.Contains(t, result, "test_data.Top")
	assert.Contains(t, result, "test_data.Bottom")
	assert.Contains(t, result, "test_data.Alternative")

	assert.Contains(t, result, "100.0%")
	assert.Contains(t, result, "50.0%")
	assert.Contains(t, result, "0.0%")

	lines := strings.Split(strings.TrimSpace(result), "\n")
	assert.Equal(t, 5, len(lines))
}
