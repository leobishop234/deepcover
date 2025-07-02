package out

import (
	"strings"
	"testing"

	"github.com/leobishop234/deepcover/src/cover"
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

	if !strings.Contains(result, "PATH") {
		t.Error("Output should contain 'PATH' header")
	}
	if !strings.Contains(result, "FUNCTION") {
		t.Error("Output should contain 'FUNCTION' header")
	}
	if !strings.Contains(result, "COVERAGE") {
		t.Error("Output should contain 'COVERAGE' header")
	}

	if !strings.Contains(result, "test_data.Top") {
		t.Error("Output should contain 'test_data.Top'")
	}
	if !strings.Contains(result, "test_data.Bottom") {
		t.Error("Output should contain 'test_data.Bottom'")
	}
	if !strings.Contains(result, "test_data.Alternative") {
		t.Error("Output should contain 'test_data.Alternative'")
	}

	if !strings.Contains(result, "100.0%") {
		t.Error("Output should contain '100.0%'")
	}
	if !strings.Contains(result, "50.0%") {
		t.Error("Output should contain '50.0%'")
	}
	if !strings.Contains(result, "0.0%") {
		t.Error("Output should contain '0.0%'")
	}

	lines := strings.Split(strings.TrimSpace(result), "\n")
	expectedLines := 5
	if len(lines) != expectedLines {
		t.Errorf("Expected %d lines, got %d", expectedLines, len(lines))
	}
}
