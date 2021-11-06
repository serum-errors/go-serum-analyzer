package analysis

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/ast/astutil"
)

type (
	taintSpreadResult struct {
		expressions        []ast.Expr             // expressions that represent the taint, or nil
		destructAssignment []*taintSpreadDestruct // taint originating from destructirung assignments, or nil
		identOutOfScope    []*ast.Ident           // every used ident that was not defined in functio scope, or nil
	}

	taintSpread struct {
		pass          *analysis.Pass
		function      *funcDefinition
		immutableType bool

		result *taintSpreadResult

		visited map[*ast.Object]struct{}
	}

	taintSpreadDestruct struct {
		position int
		target   *ast.Ident
		source   ast.Expr
	}
)

func taintSpreadForIdentOfImmutableType(pass *analysis.Pass, ident *ast.Ident, function *funcDefinition) *taintSpreadResult {
	ts := taintSpread{pass, function, true, &taintSpreadResult{}, map[*ast.Object]struct{}{}}
	ts.findSpread(ident)
	return ts.result
}

func taintSpreadForIdentAllowLeak(pass *analysis.Pass, visited map[*ast.Object]struct{}, ident *ast.Ident, function *funcDefinition) *taintSpreadResult {
	ts := taintSpread{pass, function, false, &taintSpreadResult{}, visited}
	ts.findSpread(ident)
	return ts.result
}

func (ts *taintSpread) findSpread(ident *ast.Ident) {
	if isIdentOriginOutsideFunctionScope(ts.function, ident) {
		ts.result.identOutOfScope = append(ts.result.identOutOfScope, ident)
		return
	}

	// Cannot spread taint for nil identifier.
	if ident.Name == "nil" {
		return
	}

	// Mark ident as visited to avoid revisiting it again (possibly resulting in an endless loop)
	if _, ok := ts.visited[ident.Obj]; ok {
		return
	}
	ts.visited[ident.Obj] = struct{}{}

	// Check if there can be an error codes extracted from the ident declaration statement if there is any.
	initValue := ts.findValueForIdentInValueSpec(ident)
	if initValue != nil {
		ts.processAssignedExpr(initValue)
	}

	ast.Inspect(ts.function.body(), func(node ast.Node) bool {
		// n.b., do *not* filter out *`ast.FuncLit`: statements inside closures can assign things!
		assignment, ok := node.(*ast.AssignStmt)
		if !ok {
			return true
		}

		// Look for our ident's object in the left-hand-side of the assign.
		// Either follow up on the statement at the same index in the Rhs,
		// or watch out for a shorter Rhs that's just a CallExpr (i.e. it's a destructuring assignment).
		for i, lhsEntry := range assignment.Lhs {
			lhsEntry, ok := astutil.Unparen(lhsEntry).(*ast.Ident)
			if !ok {
				continue
			}

			if lhsEntry.Obj != ident.Obj {
				continue
			}

			if len(assignment.Lhs) != len(assignment.Rhs) {
				ts.result.destructAssignment = append(ts.result.destructAssignment, &taintSpreadDestruct{i, lhsEntry, assignment.Rhs[0]})
			} else {
				ts.processAssignedExpr(assignment.Rhs[i])
			}
		}

		return true
	})
}

func (ts *taintSpread) processAssignedExpr(expr ast.Expr) {
	expr = astutil.Unparen(expr)
	ident, ok := expr.(*ast.Ident)
	if ok {
		if ident.Obj != nil && ident.Obj.Kind == ast.Var {
			ts.findSpread(ident)
			return
		}
	}
	ts.result.expressions = append(ts.result.expressions, expr)
}

// findValueForIdentInValueSpec finds the respective value for the given ident if
// the ident was declared in a ast.ValueSpec and a value was assigned at declaration.
func (ts *taintSpread) findValueForIdentInValueSpec(ident *ast.Ident) ast.Expr {
	if ident == nil || ident.Obj == nil {
		return nil
	}

	spec, ok := ident.Obj.Decl.(*ast.ValueSpec)
	if !ok || len(spec.Values) == 0 {
		return nil
	}

	for i, specIdent := range spec.Names {
		if ident.Obj == specIdent.Obj {
			if len(spec.Values) == len(spec.Names) {
				return spec.Values[i]
			} else {
				ts.result.destructAssignment = append(ts.result.destructAssignment, &taintSpreadDestruct{i, specIdent, spec.Values[0]})
				return nil
			}
		}
	}

	return nil
}
