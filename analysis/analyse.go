package analysis

import (
	"fmt"
	"go/ast"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

var Analyzer = &analysis.Analyzer{
	Name:     "rerr-lint",
	Doc:      "Checks for exhaustive handling of errors based on rerr codes.",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	nodeFilter := []ast.Node{
		(*ast.File)(nil),
		(*ast.CallExpr)(nil),
	}

	inspect.Preorder(nodeFilter, func(node ast.Node) {
		switch stmt := node.(type) {
		case *ast.File:
			fmt.Printf("%#v\n", stmt)
		case *ast.CallExpr:
			fmt.Printf("%#v\n", stmt)
			pass.Reportf(stmt.Fun.Pos(), "a call")
			fmt.Printf("%s\n", pass.Fset.PositionFor(stmt.Fun.Pos(), true))
		}
	})

	return nil, nil
}
