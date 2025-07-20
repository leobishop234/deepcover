package out

import (
	"fmt"
	"os"
	"strings"

	"github.com/leobishop234/deepcover/src/cover"
)

const coverageFormat = "%s\t\t%s\t\t%.1f%%\n"

func OutputFile(path string, coverage []cover.Coverage) error {
	coverageFile, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create coverage file: %v", err)
	}
	defer coverageFile.Close()

	coverageFile.WriteString(formatFile(coverage))
	return nil
}

func formatFile(coverage []cover.Coverage) string {
	var str strings.Builder
	for _, cover := range coverage {
		str.WriteString(fmt.Sprintf(coverageFormat, cover.Name, cover.Path, cover.Coverage))
	}

	return str.String()
}
