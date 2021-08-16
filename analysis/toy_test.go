package analysis

import (
	"fmt"
	"go/ast"
	"os"
	"reflect"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

var ToyAnalyzer = &analysis.Analyzer{
	Name:     "demo",
	Doc:      "Checks for exhaustive handling of errors based on ree codes.",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      runToy,
}

func runToy(pass *analysis.Pass) (interface{}, error) {
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
			ast.Fprint(os.Stdout, pass.Fset, stmt.Body, ast.NotNilFilter)
		}
	})

	// seems to be roughly:
	// - look for `*ast.ReturnStmt`
	// - pluck out `.Results[relevantone].(*ast.Ident).Obj` -- this is a pointer
	//   - ... i guess whole expressions can be in here too instead of just an Ident, so have fun with that.  Recursion on ast.Expr, maybe, is the right thing?
	// - look for any `*ast.AssignStmt`, look for their `.Lhs`, these should all be `*ast.Ident`, look at the `.Obj` -- is it the familiar one?  Okay, follow the matching `.Rhs`.
	//   - yes, indeed -- this `.Obj` stuff is tracking the actual origin of the var, so anything that assigns to this, we found it.
	//   - ... i guess probably the `.Lhs` can have a struct var ref or something too probably, not sure if that's relevant to us though.
	//     - these are just called `*ast.SelectorExpr`, I think.

	// so we can track everything that assigns to a var (ignoring conditions) fairly easily, apparently.
	// (i haven't checked closures yet, but kinda assuming at this point.)  (... okay, now i have, yes, they are seen too.  convenient.)
	// the one big obvious limitation here is... if you assign something to a var, and then try to filter it and assign back to the same var, i can't easily see that.
	// (which is awfully unfortunate, because the demo I wrote earlier definitely tries to do exactly that, heh.  (maybe it's not good patterns anyway; unknown.))
	// i'm still trying to refrain from needing full on symbolic execution or ssa here.
	//
	// i guess if you are always doing the retagging right before returning (like, literally in the same expression), this is clear and fine.
	// maybe we can look back a very limited amount before a return statement, in linearly preceding statements?... and as long as we can find a retagging that's unconditionally "before", it counts?  (so, pop up, and step back, but never deeper, more or less should do it?  technically you'd miss an unconditional closure that way, but whatever.)

	// remember there's also a second half: walking backward from return statements is for checking what the current function can yield.
	// there's also the need to check that it handles (or at least acknowledges) whatever its child calls can return.  that flows the opposite direction.

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
