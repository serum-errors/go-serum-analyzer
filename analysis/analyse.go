package analysis

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"

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
		logf("func %q returns an error interface (type name: %q)\n", funcDecl.Name.Name, typ)
		funcsToAnalyse = append(funcsToAnalyse, funcDecl)
	})
	logf("%d functions in this package return errors and will be analysed.\n\n", len(funcsToAnalyse))

	// First output: warn directly about any functions that are exported
	// if they return errors, but don't declare error codes in their docs.
	// TODO

	return nil, nil
}

var tError = types.NewInterfaceType([]*types.Func{
	types.NewFunc(token.NoPos, nil, "Error", types.NewSignature(nil, nil, types.NewTuple(types.NewVar(token.NoPos, nil, "", types.Typ[types.String])), false)),
}, nil).Complete()

func init() {
	//fmt.Printf("%v\n\n", tError)
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
