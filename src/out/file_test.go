package out

import (
	"os"
	"testing"

	"github.com/leobishop234/deepcover/src/cover"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var fileTestCoverage = cover.Result{
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
	expected := `package.Function1		example/path/file1.go		100.00%
package.Function2		example/path/file2.go		50.00%
package.Function3		example/path/file3.go		0.00%
Total: 50.00%
`

	got := formatFile(fileTestCoverage)

	assert.Equal(t, expected, got)
}
