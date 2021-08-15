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

var VerifyAnalyzer = &analysis.Analyzer{
	Name:     "ree-verify",
	Doc:      "Checks that any function that has a ree-style docstring enumerating error codes is telling the truth.",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      runVerify,
}

// FUTURE: may add another analyser that is "ree-exhaustive".

func runVerify(pass *analysis.Pass) (interface{}, error) {
	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	// First pass: let's just see what error types we might need to reason about.
	// We'll do this by looking at... every type seen by the analysis pass.  Yep.
	// (I can't see any easier way to do this, unfortunately.)
	// ... oh great, these pointers aren't even unique.  For real?
	// ........ https://godoc.org/golang.org/x/tools/go/types/typeutil apparently has a feature for this.
	allTypes := map[types.Type]struct{}{}
	for _, typ := range pass.TypesInfo.Types {
		allTypes[typ.Type] = struct{}{}
	}
	for typ := range allTypes {
		fmt.Printf("%v\n", typ)
	}
	fmt.Print("\n\n\n")

	// We only need to see function declarations at first; we'll recurse ourselves within there.
	nodeFilter := []ast.Node{
		(*ast.FuncDecl)(nil),
	}

	// Let's look only at functions that return errors;
	// and furthermore, errors as their last result (that's a normal enough convention, isn't it?).
	inspect.Preorder(nodeFilter, func(node ast.Node) {
		funcDecl := node.(*ast.FuncDecl)
		resultsList := funcDecl.Type.Results
		if resultsList == nil {
			return
		}
		for _, resultFieldClause := range resultsList.List {
			typ := pass.TypesInfo.Types[resultFieldClause.Type].Type
			fmt.Printf("%v -- %v\n", typ, types.Implements(typ, tError))
			// TODO ....
		}
	})

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
