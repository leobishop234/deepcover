# Deepcover

Deepcover is a Go CLI tool to calculate deep code coverage for your go tests by analysing a function's downstream dependencies across packages using the [Class Hierarchy Analysis](https://pkg.go.dev/golang.org/x/tools/go/callgraph/cha) algorithm.

## Installation
```bash
go install github.com/yourusername/deepcover@latest
```

Or build from source:

```bash
git clone https://github.com/yourusername/deepcover
cd deepcover
go build -o deepcover
```

## Usage

```bash
deepcover [flags] <package-path>
```

### Flags

- `-run string`: Unanchored regular expression that matches target test names, if not provided defaults to all tests in the target package
- `-o string`: Output file path (if not provided, deepcover outputs to terminal)

### Examples

Calculate deep coverage for all tests in the target package:
```bash
deepcover ./mypackage
```

Calculate deep coverage for a specific test:
```bash
deepcover -run TestUser ./mypackage
```

Calculate deep coverage for all tests matching a specified pattern:
```bash
deepcover -run "Test.*Integration" ./mypackage
```

Save deep coverage statistics to a target file.
```bash
deepcover -run "Test.*" -o coverage.txt ./mypackage
```

## Output Format

Deepcover outputs a table showing:
- **PATH**: The file path and line number of the function
- **FUNCTION**: The function name
- **COVERAGE**: The percentage of the function covered by the tests

Example output:
```
PATH                                          FUNCTION       COVERAGE
---------------------------------------------------------------------
deepcover/testexample/example.go:5:           Top            100.0%  
deepcover/testexample/example.go:9:           Bottom         100.0%  
deepcover/testexample/interface.go:9:         newInterface   100.0%  
deepcover/testexample/interface.go:15:        Method         66.7%   
deepcover/testexample/subpkg/subtest.go:12:   SubPkg         75.0%   
```
## Requirements

- Go 1.24.2 or later
- The target package must be part of a Go module

## License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.

## Further Work

Deepcover is currently in an MVP state, further work is expected, and any contributions are welcome.

Intended work includes:
- Calculation of total coverage of identified dependencies.
- Support for the [Rapid Type Analysis](https://pkg.go.dev/golang.org/x/tools/go/callgraph/rta) and [Variable Type Analysis](https://pkg.go.dev/golang.org/x/tools/go/callgraph/vta) callgraph algorithms.
- Support for targeting multiple packages simultaneously.