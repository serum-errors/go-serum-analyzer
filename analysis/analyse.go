package analysis

import (
	"fmt"
	"go/ast"
	"reflect"
	"strings"

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
			fmt.Printf("\tfunc name: %#v\n", stmt.Name.Name)
			if stmt.Recv != nil { // not really sure why there's a list here, it's either nil or it's one element.
				fmt.Printf("\trecv: %s\n", quickString(stmt.Recv))
			}
			fmt.Printf("\tfunc sig: %s -> %s\n", quickString(stmt.Type.Params), quickString(stmt.Type.Results))
			fmt.Printf("\tbody: %#v\n", stmt.Body.List)
		}
	})

	return nil, nil
}

// quickString spits out a string for something quickly and briefly, and not entirely recursively.
//
// you can use `ast.Fprint(os.Stdout, pass.Fset, stmt, ast.NotNilFilter)` if you're really wanting a mouthful,
// but it's a truly huge output, includes lots of position info you'd probably find redundant, etc.
func quickString(x interface{}) string {
	if reflect.ValueOf(x).IsNil() {
		return fmt.Sprintf("%T(nil)", x)
	}
	switch y := x.(type) {
	case *ast.FieldList:
		var sb strings.Builder
		sb.WriteString("(")
		for _, field := range y.List {
			for _, name := range field.Names { // mind: you're experiencing AST here: it can have multiple names per type.
				sb.WriteString(name.Name)
				sb.WriteString(" ")
				sb.WriteString(quickString(field.Type)) // this can be a whole heckin expression, technically.  (and often is, because commonly it has at least a "StarExpr" kicking it off.)
				sb.WriteString(",")
			}
			if len(field.Names) == 0 {
				// or, it can have no name at all, just a type.
				sb.WriteString("_ ")
				sb.WriteString(quickString(field.Type))
				sb.WriteString(",")
			}
		}
		sb.WriteString(")")
		return sb.String()
	case *ast.StarExpr:
		return fmt.Sprintf("*(%v)", y.X)
	case fmt.Stringer:
		return y.String()
	case fmt.GoStringer:
		return y.GoString()
	default:
		return fmt.Sprintf("%#v", x)
	}
}
