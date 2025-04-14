// Package metcolanalyzers contains various custom static analysers.
package metcolanalyzers

import (
	"errors"
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

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
