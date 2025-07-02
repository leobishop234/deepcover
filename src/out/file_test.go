package out

import (
	"os"
	"strings"
	"testing"

	"github.com/leobishop234/deepcover/src/cover"
	"github.com/stretchr/testify/require"
)

func TestSaveFile(t *testing.T) {
	coverage := []cover.Coverage{
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

	temp, err := os.CreateTemp(t.TempDir(), "*.coverage")
	require.NoError(t, err)
	defer temp.Close()

	require.NoError(t, SaveFile(temp.Name(), coverage))

	gotBytes, err := os.ReadFile(temp.Name())
	require.NoError(t, err)
	require.Equal(t, coverageArrayToString(coverage), string(gotBytes))
}

func coverageArrayToString(coverage []cover.Coverage) string {
	lines := make([]string, len(coverage))
	for i, c := range coverage {
		lines[i] = coverageString(c)
	}
	return strings.Join(lines, "")
}
