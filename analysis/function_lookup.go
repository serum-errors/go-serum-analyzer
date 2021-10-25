package analysis

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
	"golang.org/x/tools/go/types/typeutil"
)

// funcLookup allows the performant lookup of function and method declarations in the current package by name,
// and the lookup of cached error codes and affectors for function declarations.
type funcLookup struct {
	functions  map[string]*ast.FuncDecl   // Mapping Function Names to Declarations
	methods    map[string][]*ast.FuncDecl // Mapping Method Names to Declarations (Multiple Possible per Name)
	methodSet  typeutil.MethodSetCache
	foundCodes map[funcDeclOrLit]CodeSet // Mapping Function Declarations and Function Literals to cached error codes
}

func newFuncLookup() *funcLookup {
	return &funcLookup{
		map[string]*ast.FuncDecl{},
		map[string][]*ast.FuncDecl{},
		typeutil.MethodSetCache{},
		map[funcDeclOrLit]CodeSet{},
	}
}

// collectFunctions creates a funcLookup using the given analysis object.
func collectFunctions(pass *analysis.Pass) *funcLookup {
	result := newFuncLookup()
	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	// We only need to see function declarations at first; we'll recurse ourselves within there.
	nodeFilter := []ast.Node{
		(*ast.FuncDecl)(nil),
	}

	inspect.Nodes(nodeFilter, func(node ast.Node, _ bool) bool {
		funcDecl := node.(*ast.FuncDecl)

		// Check if it's a function or a method and add accordingly.
		if !isMethod(funcDecl) {
			result.functions[funcDecl.Name.Name] = funcDecl
		} else {
			result.methods[funcDecl.Name.Name] = append(result.methods[funcDecl.Name.Name], funcDecl)
		}

		// Never recurse into the function bodies
		return false
	})

	return result
}

// forEach traverses all the functions and methods in the lookup,
// and applies the given function f to every ast.FuncDecl.
func (lookup *funcLookup) forEach(f func(*ast.FuncDecl)) {
	for _, funcDecl := range lookup.functions {
		f(funcDecl)
	}

	for _, methods := range lookup.methods {
		for _, funcDecl := range methods {
			f(funcDecl)
		}
	}
}

// searchMethodType searches for method in the type information using receiver type and method name.
func (lookup *funcLookup) searchMethodType(pass *analysis.Pass, receiver types.Type, methodName string) *types.Selection {
	methodSet := lookup.methodSet.MethodSet(receiver)
	searchedMethodType := methodSet.Lookup(pass.Pkg, methodName)

	if searchedMethodType == nil {
		// No methods were found for T
		// Search methods for *T if T is not already a pointer.
		_, ok := receiver.(*types.Pointer)
		if !ok {
			methodSet = lookup.methodSet.MethodSet(types.NewPointer(receiver))
			searchedMethodType = methodSet.Lookup(pass.Pkg, methodName)
		}
	}

	return searchedMethodType
}

// searchMethod tries to find the correct function declaration for a method given the receiver type and method name.
func (lookup *funcLookup) searchMethod(pass *analysis.Pass, receiver types.Type, methodName string) *ast.FuncDecl {
	methods, ok := lookup.methods[methodName]
	if !ok || len(methods) == 0 {
		// Return early if there is no method in the current package with the given name
		return nil
	}

	searchedMethodType := lookup.searchMethodType(pass, receiver, methodName)
	if searchedMethodType == nil {
		// Return early if there is no method matching receiver and name
		return nil
	}

	// Method we're looking for exists in the current package, we only need to find the right declaration
	for _, method := range methods {
		methodObj := pass.TypesInfo.ObjectOf(method.Name)
		if searchedMethodType.Obj() == methodObj {
			return method
		}
	}

	return nil
}
