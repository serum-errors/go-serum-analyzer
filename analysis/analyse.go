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

	for _, funcDecl := range funcsToAnalyse {
		claimedCodes := funcClaims[funcDecl]
		if claimedCodes == nil {
			continue
		}
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
				_ = stmt.Results
			}
			return true
		})
	}

	return nil, nil
}

var tError = types.NewInterfaceType([]*types.Func{
	types.NewFunc(token.NoPos, nil, "Error", types.NewSignature(nil, nil, types.NewTuple(types.NewVar(token.NoPos, nil, "", types.Typ[types.String])), false)),
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

// findErrorTypes looks at every function in the package,
// and then at each of their return types,
// and determines if they match the ree.Error interface,
// and if they do, adds them to the list to be returned.
//
// This should of course identify ree.ErrorStruct itself,
// but may also identify other types in other libraries that match.
func findErrorTypes() {

}

// checkErrorTypeHasLegibleCode makes sure that the `Code() string` function
// on a type either returns a constant or a single struct field.
// If you want to write your own ree.Error, it should be this simple.
func checkErrorTypeHasLegibleCode() { // probably should return a lookup function.

}
