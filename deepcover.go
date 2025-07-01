package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/leobishop234/deepcover/src/cover"
	"github.com/leobishop234/deepcover/src/out"
)

func main() {
	var target string
	var output string

	flag.StringVar(&target, "run", "Test", "Unanchored regular expression that matches target test names")
	flag.StringVar(&output, "o", "", "Output file path")

	flag.Parse()

	args := flag.Args()
	if len(args) != 1 {
		fmt.Fprintf(os.Stderr, "Error: Expected path to target package as argument\n")
		os.Exit(1)
	}
	pkgPath := args[0]

	if err := run(pkgPath, target, output); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run(pkgPath, target, output string) error {
	if pkgPath == "" {
		return fmt.Errorf("pkg path is required")
	}

	coverage, err := cover.Deepcover(pkgPath, target)
	if err != nil {
		return fmt.Errorf("failed to get dependencies: %v", err)
	}

	if output != "" {
		if err := out.SaveFile(output, coverage); err != nil {
			return fmt.Errorf("failed to save coverage to file: %v", err)
		}
	} else {
		out.PrintCoverage(coverage)
	}

	return nil
}
