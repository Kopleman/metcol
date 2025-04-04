//nolint:godot // dunno why it not like first comment block
/*
* Package  implements an extended static analysis tool for Go.

It combines:

1. Standard analyzers from golang.org/x/tools/go/analysis/passes
2. All SA-class analyzers from staticcheck.io
3. Selected analyzers from other staticcheck classes (ST, QF)
4. Custom analyzer that prohibits os.Exit in main function

## Installation

Ensure Go modules are initialized, then install dependencies:

go get -u golang.org/x/tools/go/analysis/passes/...
go get -u honnef.co/go/tools/staticcheck

## Building

go build -o staticlint cmd/staticlint/main.go

## Usage

Analyze current project:

./staticlint ./...

Analyze specific package:

./staticlint path/to/package

## Included Analyzers

### Standard Analyzers
- asmdecl    - check assembly declarations
- assign     - detect useless assignments
- atomic     - check for common atomic mistakes
- bools      - check for common boolean mistakes
- buildtag   - validate build tags
- cgocall    - detect common cgo mistakes
- composite  - check for unkeyed struct literals
- copylock   - check for locks erroneously passed by value
- httpresponse - check for mistakes using HTTP responses

### Staticcheck SA-class
- SAxxxx series - detect common bugs:
  - SA1000 - invalid regex syntax
  - SA1012 - invalid time format
  - ... [other SA analyzers]

### Other staticcheck classes
- ST1000 - style checks for naming conventions
- QF1000 - suggested quick fixes

### Custom Analyzers
- noosexit - prohibits direct os.Exit calls in main.main

## Custom noosexit Analyzer

Purpose: Prevent usage of os.Exit in main function of main package.

Mechanism:
1. Checks if analyzed package is "main"
2. Locates main function
3. Verifies all function calls within main
4. Reports any os.Exit calls with exact position

Example violation:

	func main() {
	    os.Exit(1) // diagnostic: os.Exit call forbidden in main.main
	}

Recommendations:
- Return error codes from main instead
- Move exit logic to separate functions
- Handle errors through logging and status returns
*/
package main

import (
	"errors"
	"go/ast"
	"go/types"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/asmdecl"
	"golang.org/x/tools/go/analysis/passes/assign"
	"golang.org/x/tools/go/analysis/passes/atomic"
	"golang.org/x/tools/go/analysis/passes/bools"
	"golang.org/x/tools/go/analysis/passes/buildtag"
	"golang.org/x/tools/go/analysis/passes/cgocall"
	"golang.org/x/tools/go/analysis/passes/composite"
	"golang.org/x/tools/go/analysis/passes/copylock"
	"golang.org/x/tools/go/analysis/passes/httpresponse"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
	"honnef.co/go/tools/quickfix/qf1001"
	"honnef.co/go/tools/staticcheck"
	"honnef.co/go/tools/stylecheck/st1000"
)

func main() {
	analyzers := []*analysis.Analyzer{
		// Standard analyzers
		asmdecl.Analyzer,
		assign.Analyzer,
		atomic.Analyzer,
		bools.Analyzer,
		buildtag.Analyzer,
		cgocall.Analyzer,
		composite.Analyzer,
		copylock.Analyzer,
		httpresponse.Analyzer,
	}

	// Staticcheck SA-class analyzers
	for _, a := range staticcheck.Analyzers {
		if strings.HasPrefix(a.Analyzer.Name, "SA") {
			analyzers = append(analyzers, a.Analyzer)
		}
	}

	// Other staticcheck analyzers
	analyzers = append(analyzers,
		st1000.SCAnalyzer.Analyzer, // Naming conventions
		qf1001.SCAnalyzer.Analyzer, // Quick fixes
		NoOsExitAnalyzer,           // Custom analyzer
	)

	multichecker.Main(analyzers...)
}

// NoOsExitAnalyzer prohibits direct calls to os.Exit in main.main
// The analyzer checks:
// - Package is named "main"
// - Context is within main function
// - os.Exit usage through import verification.
var NoOsExitAnalyzer = &analysis.Analyzer{
	Name:     "noosexit",
	Doc:      "forbid direct calls to os.Exit in main.main",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      runNoOsExit,
}

func runNoOsExit(pass *analysis.Pass) (interface{}, error) {
	if pass.Pkg.Name() != "main" {
		return nil, nil //nolint:nilnil // specific case
	}

	insp, ok := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	if !ok {
		return errors.New("invalid type conversion"), nil
	}
	nodeFilter := []ast.Node{
		(*ast.FuncDecl)(nil),
		(*ast.CallExpr)(nil),
	}

	var inMain bool

	insp.Preorder(nodeFilter, func(n ast.Node) {
		switch node := n.(type) {
		case *ast.FuncDecl:
			inMain = node.Name.Name == "main"
		case *ast.CallExpr:
			if !inMain {
				return
			}

			sel, ok := node.Fun.(*ast.SelectorExpr)
			if !ok {
				return
			}

			pkg, ok := sel.X.(*ast.Ident)
			if !ok {
				return
			}

			if pkg.Name == "os" && sel.Sel.Name == "Exit" {
				if obj, ok := pass.TypesInfo.Uses[pkg].(*types.PkgName); ok && obj.Imported().Path() == "os" {
					pass.Reportf(node.Pos(), "os.Exit call forbidden in main.main")
				}
			}
		}
	})

	return nil, nil //nolint:nilnil // specific case
}
