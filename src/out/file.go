package out

import (
	"fmt"
	"os"

	"github.com/leobishop234/deepcover/src/cover"
)

const coverageFormat = "%s\t\t%s\t\t%.1f%%\n"

func SaveFile(path string, coverage []cover.Coverage) error {
	coverageFile, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create coverage file: %v", err)
	}
	defer coverageFile.Close()

	for _, funcCoverage := range coverage {
		_, err := coverageFile.WriteString(coverageString(funcCoverage))
		if err != nil {
			return fmt.Errorf("failed to write coverage to file: %v", err)
		}
	}

	return nil
}

func coverageString(coverage cover.Coverage) string {
	return fmt.Sprintf(coverageFormat, coverage.Path, coverage.Name, coverage.Coverage)
}
