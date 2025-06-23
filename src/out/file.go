package out

import (
	"deepcover/src/cover"
	"fmt"
	"os"
)

func SaveFile(path string, coverage []cover.Coverage) error {
	coverageFile, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create coverage file: %v", err)
	}
	defer coverageFile.Close()

	for _, funcCoverage := range coverage {
		_, err := coverageFile.WriteString(fmt.Sprintf("%s\t\t%s\t\t%.1f%%\n", funcCoverage.Path, funcCoverage.Name, funcCoverage.Coverage))
		if err != nil {
			return fmt.Errorf("failed to write coverage to file: %v", err)
		}
	}

	return nil
}
