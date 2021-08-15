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
		(*ast.FuncDecl)(nil),
		(*ast.FuncLit)(nil),
	}

	inspect.Preorder(nodeFilter, func(node ast.Node) {
		fmt.Printf("node: %T -- pos: %s -- content: %#v\n", node, pass.Fset.PositionFor(node.Pos(), true), node)
		switch stmt := node.(type) {
		case *ast.File:
		case *ast.CallExpr:
		case *ast.FuncDecl: // n.b. does not include inlines -- that's a "FuncLit".  *does* include methods, though.
			fmt.Printf("\t%#v\n", stmt.Body.List)
		}
	})

	return nil, nil
}
