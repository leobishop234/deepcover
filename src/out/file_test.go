package out

import (
	"os"
	"testing"

	"github.com/leobishop234/deepcover/src/cover"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var fileTestCoverage = []cover.Coverage{
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

func TestOutputFile(t *testing.T) {
	temp, err := os.CreateTemp(t.TempDir(), "*.coverage")
	require.NoError(t, err)
	defer temp.Close()

	require.NoError(t, OutputFile(temp.Name(), fileTestCoverage))

	gotBytes, err := os.ReadFile(temp.Name())
	require.NoError(t, err)
	assert.Equal(t, formatFile(fileTestCoverage), string(gotBytes))
}

func TestFormatFile(t *testing.T) {
	expected := `test_data.Top		test_data/example.go		100.0%
test_data.Bottom		test_data/example.go		50.0%
test_data.Alternative		test_data/example.go		0.0%
`

	got := formatFile(fileTestCoverage)

	assert.Equal(t, expected, got)
}
