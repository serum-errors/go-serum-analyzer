package analysis

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

var logf = fmt.Printf

var VerifyAnalyzer = &analysis.Analyzer{
	Name:     "ree-verify",
	Doc:      "Checks that any function that has a ree-style docstring enumerating error codes is telling the truth.",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      runVerify,
}

// FUTURE: may add another analyser that is "ree-exhaustive".

func runVerify(pass *analysis.Pass) (interface{}, error) {
	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	// We only need to see function declarations at first; we'll recurse ourselves within there.
	nodeFilter := []ast.Node{
		(*ast.FuncDecl)(nil),
	}

	// Let's look only at functions that return errors;
	// and furthermore, errors as their last result (that's a normal enough convention, isn't it?).
	//
	// Returning more than one error will result in anything but the last one not being analysed.
	// Returning an error in any result field but the last one will result in it not being analysed.
	//
	// We'll actually look for anything that _implements_ `error` (!), not just the literal type.
	// Sometimes these will also, furthermore, perhaps implement our own extended error interface...
	// but if so, that's something we'll look into more later, not right now.
	var funcsToAnalyse []*ast.FuncDecl
	inspect.Preorder(nodeFilter, func(node ast.Node) {
		funcDecl := node.(*ast.FuncDecl)
		resultsList := funcDecl.Type.Results
		if resultsList == nil {
			return
		}
		lastResult := resultsList.List[len(resultsList.List)-1]
		typ := pass.TypesInfo.Types[lastResult.Type].Type
		if !types.Implements(typ, tError) {
			return
		}
		logf("function %q returns an error interface (type name: %q)\n", funcDecl.Name.Name, typ)
		funcsToAnalyse = append(funcsToAnalyse, funcDecl)
	})
	logf("%d functions in this package return errors and will be analysed.\n\n", len(funcsToAnalyse))

	// First output: warn directly about any functions that are exported
	// if they return errors, but don't declare error codes in their docs.
	// Also: because we have to do the whole parse for docstrings already,
	// remember the error codes for the funcs that do have them:
	// those are what we'll look at for the remaining analysis.
	funcClaims := map[*ast.FuncDecl][]string{}
	for _, funcDecl := range funcsToAnalyse {
		codes, err := findErrorDocs(funcDecl)
		if err != nil {
			pass.Reportf(funcDecl.Pos(), "function %q has odd docstring: %s", funcDecl.Name.Name, err)
			continue
		}
		if codes == nil {
			if funcDecl.Name.IsExported() {
				pass.Reportf(funcDecl.Pos(), "function %q is exported, but does not declare any error codes", funcDecl.Name.Name)
			}
		} else {
			funcClaims[funcDecl] = codes
			logf("function %q declares error codes %s\n", funcDecl.Name.Name, codes)
		}
	}
	logf("%d functions in this package return errors and declared codes for them, and will be further analysed.\n\n", len(funcClaims))

	// Okay -- let's look at the functions that have made claims about their error codes.
	// We'll explore deeply to find everything that can actually affect their error return value.
	// When we reach data initialization... we look at if those types implement coded errors, and try to figure out what their actual code value is.
	// When we reach other function calls that declare their errors, that's good enough info (assuming they're also being checked for truthfulness).
	// Anything else is trouble.
	for _, funcDecl := range funcsToAnalyse {
		claimedCodes := funcClaims[funcDecl]
		if claimedCodes == nil {
			continue
		}
		affectOrigins := findAffectorsOfErrorReturnInFunc(pass, funcDecl)
		logf("trace found these origins of error data...\n")
		for _, aff := range affectOrigins {
			logf(" - %s -- %s -- %v\n", pass.Fset.PositionFor(aff.Pos(), true), aff, checkErrorTypeHasLegibleCode(pass, aff))
		}
		logf("end of found origins.\n\n")

	}

	return nil, nil
}

var tError = types.NewInterfaceType([]*types.Func{
	types.NewFunc(token.NoPos, nil, "Error", types.NewSignature(nil, nil, types.NewTuple(types.NewVar(token.NoPos, nil, "", types.Typ[types.String])), false)),
}, nil).Complete()

var tReeError = types.NewInterfaceType([]*types.Func{
	types.NewFunc(token.NoPos, nil, "Error", types.NewSignature(nil, nil, types.NewTuple(types.NewVar(token.NoPos, nil, "", types.Typ[types.String])), false)),
	types.NewFunc(token.NoPos, nil, "Code", types.NewSignature(nil, nil, types.NewTuple(types.NewVar(token.NoPos, nil, "", types.Typ[types.String])), false)),
}, nil).Complete()

func init() {
	//fmt.Printf("%v\n\n", tError)
}

// findErrorDocs looks at the doc comments on a function,
// tries to parse out error code declarations that we can recognize,
// and returns the error code strings from that.
//
// The declaration format is:
//   - strip a leading "^//" if present.
//   - strip any more leading whitespace.
//   - a line that is exactly "Errors:" starts a declaration block.
//   - exactly one blank line must follow, or it's a bad format.
//   - the next line must match "^- (.*) --", and the captured group is an error code.
//     - note that this is after leading whitespace strip.  (probably you should indent these, for readability.)
//     - for simplier parsing, any line that starts with "- " will be slurped,
//       and we'll consider it an error if the rest of the pattern doesn't follow.
//     - the capture group can be stripped for whitespace again.  (perhaps the author wanted to align things.)
//   - this may repeat.  if lines do not start that that pattern, they are skipped.
//      - note that the same code may appear multiple times.  this is acceptable, and should be deduplicated.
//   - when there's another fully blank line, the parse is ended.
// This format happens to be amenable to letting you write the closest thing godocs have to a list.
// (You should probably indent things "enough" to make that render right, but we're not checking that here right now.)
//
// If there are no error declarations, (nil, nil) is returned.
// If there's what looks like an error declaration, but funny looking, an error is returned.
func findErrorDocs(funcDecl *ast.FuncDecl) ([]string, error) {
	if funcDecl.Doc == nil {
		return nil, nil
	}
	var parsing, needBlankLine bool
	var codes []string
	seen := map[string]struct{}{}
	for _, line := range strings.Split(funcDecl.Doc.Text(), "\n") {
		line := strings.TrimSpace(line)
		switch {
		case needBlankLine && line != "":
			return nil, fmt.Errorf("need a blank line after the 'Errors:' block indicator")
		case needBlankLine && line == "":
			needBlankLine = false
		case line == "Errors:" && parsing == false:
			parsing = true
			needBlankLine = true
		case line == "Errors:" && parsing == true:
			return nil, fmt.Errorf("repeated 'Errors:' block indicator")
		case parsing == true && strings.HasPrefix(line, "- "):
			end := strings.Index(line, " --")
			if end == -1 {
				return nil, fmt.Errorf("mid block, a line leading with '- ' didnt contain a '--' to mark the end of the code name")
			}
			if end < 2 {
				return nil, fmt.Errorf("an error code can't be purely whitespace")
			}
			code := line[2:end]
			code = strings.TrimSpace(code)
			if code == "" {
				return nil, fmt.Errorf("an error code can't be purely whitespace")
			}
			if _, exists := seen[code]; !exists {
				seen[code] = struct{}{}
				codes = append(codes, code)
			}
		case parsing == true && line == "":
			return codes, nil
		}
	}
	return codes, nil
}

// findAffectorsInFunc looks up what can affect the given expression
// (which, generally, can be anything you'd expect to see in a ReturnStmt -- so, variables, unaryExpr, a bunch of things...),
// and recurses in this until it hits either the creation of a value,
// or function call boundaries (`*ast.CallExpr`).
//
// So, it'll follow any number of assignment statements, for example;
// as it does so, it'll totally disregarding logical branching,
// instead using a very basic model of taint: just marking anything that can ever possibly touch the variable.
//
func findAffectorsInFunc(expr ast.Expr, within *ast.FuncDecl) (result []ast.Expr) {
	switch exprt := expr.(type) {
	case *ast.CallExpr: // These are a boundary condition, so that's short and sweet.
		return []ast.Expr{expr}
	case *ast.Ident: // Lovely!  These are easy.  (Although likely to have significant taint spread.)
		// Look for for `*ast.AssignStmt` in the function that could've affected this.
		ast.Inspect(within, func(node ast.Node) bool {
			// n.b., do *not* filter out *`ast.FuncLit`: statements inside closures can assign things!
			switch stmt2 := node.(type) {
			case *ast.AssignStmt:
				// Look for our ident's object in the left-hand-side of the assign.
				// Either follow up on the statement at the same index in the Rhs,
				// or watch out for a shorter Rhs that's just a CallExpr (i.e. it's a destructuring assignment).
				for i, clause := range stmt2.Lhs {
					switch clauset := clause.(type) {
					case *ast.Ident:
						if clauset.Obj == exprt.Obj {
							if len(stmt2.Lhs) > len(stmt2.Rhs) {
								// Destructuring mode.
								// We're going to make some crass simplifications here, and say... if this is anything other than the last arg, you're not supported.
								if i != len(stmt2.Lhs)-1 {
									panic("unsupported: tracking through a call that's a destructuring assignment and has the error as non-last return")
								}
								// Because it's a CallExpr, we're done here: this is part of the result.
								switch stmt2.Rhs[0].(type) {
								case *ast.CallExpr:
									result = append(result, stmt2.Rhs[0])
								default:
									panic("what?")
								}
							} else {
								result = append(result, findAffectorsInFunc(stmt2.Rhs[i], within)...)
							}
						}
					case *ast.SelectorExpr:
						logf("findAffectorsInFunc is looking at an assignment inside a value of interest?  fun\n")
					}
				}
			}
			return true
		})
	case *ast.UnaryExpr:
		return findAffectorsInFunc(exprt.X, within)
	case *ast.CompositeLit: // Actual value creation!
		return []ast.Expr{expr} // TODO we're almost there?  would need to apply checkErrorTypeHasLegibleCode on whatever type this is... but it is at least guaranteed to be a value matching `error`.
	case *ast.BasicLit: // Actual value creation!
		return []ast.Expr{expr} // TODO we're almost there?  checkErrorTypeHasLegibleCode needed again.
	default:
		logf(":: findAffectorsInFunc does not yet handle %#v\n", expr)
	}
	return
}

func findAffectorsOfErrorReturnInFunc(pass *analysis.Pass, funcDecl *ast.FuncDecl) (affs []ast.Expr) {
	// TODO this should probably be approximately a good point for memoization?
	ast.Inspect(funcDecl, func(node ast.Node) bool {
		switch stmt := node.(type) {
		case *ast.FuncLit:
			return false // We don't want to see return statements from in a nested function right now.
		case *ast.ReturnStmt:
			// TODO stmt.Results can also be nil, in which case you have to look back at vars in the func sig.
			logf("function %q has a return statement: %s\n", funcDecl.Name.Name, stmt.Results)
			// This can go a lot of ways:
			// - You can have a plain `*ast.Ident` (aka returning a variable).
			// - You can have an `*ast.SelectorExpr` (returning a variable from in a structure).
			// - You can have an `*ast.CallExpr` (aka returning the result of a function call).
			// - You can have an `*ast.UnaryExpr` (probably about to be an '&' and then a structure literal, but could be other things too...).
			// - This is probably not an exhaustive list...

			lastResult := stmt.Results[len(stmt.Results)-1]
			affs = append(affs, findAffectors(pass, lastResult, funcDecl)...)
		}
		return true
	})
	return
}

// findAffectors applies findAffectorsInFunc, and then _keeps going_...
// until it's resolved everything into one of:
//  - value creation,
//  - a CallExpr that targets another function that has declared error codes (yay!),
//  - a CallExpr that crosses package boundaries,
//  - a CallExpr that's an interface (we can't really look deeper than that),
//  - a CallExpr it's seen before,
//  - ... I think that's it?
//
// For the first two: we're happy: we can analyse this func completely.
// Encountering any of the others means we've found a source of unknowns.
//
func findAffectors(pass *analysis.Pass, expr ast.Expr, startingFunc *ast.FuncDecl) (result []ast.Expr) {
	stepResults := findAffectorsInFunc(expr, startingFunc)
	for _, x := range stepResults {
		switch exprt := x.(type) {
		case *ast.CallExpr: // Alright, let's goooooo
			logf("fun expr time: %#v ... %#v\n", exprt.Fun, pass.TypesInfo.Types[exprt.Fun])
			switch funst := exprt.Fun.(type) {
			case *ast.Ident: // this is what calls in your own package look like. // TODO and dot-imported, I guess.  Yeesh.
				calledFunc := funst.Obj.Decl.(*ast.FuncDecl)
				result = append(result, findAffectorsOfErrorReturnInFunc(pass, calledFunc)...)
			case *ast.SelectorExpr: // this is what calls to other packages look like.
				logf("todo: findAffectors doesn't yet search beyond selector expressions %#v\n", funst)
			}
		case *ast.CompositeLit, *ast.BasicLit:
			result = append(result, x)
		}
	}
	return
}

// checkErrorTypeHasLegibleCode makes sure that the `Code() string` function
// on a type either returns a constant or a single struct field.
// If you want to write your own ree.Error, it should be this simple.
func checkErrorTypeHasLegibleCode(pass *analysis.Pass, seen ast.Expr) bool { // probably should return a lookup function.
	typ := pass.TypesInfo.Types[seen].Type
	// FIXME absolutely wild, assuming that we've got a literal here.  sometimes we have calls here, or will when we make more fun test packages or fix some of the other todos.
	typ = types.NewPointer(typ) // FIXME, terrible hack.  quietly throwing away all the unarys left us in an awkward position.
	return types.Implements(typ, tReeError)
}
