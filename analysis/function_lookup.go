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
	functions       map[string]*ast.FuncDecl   // Mapping Function Names to Declarations
	methods         map[string][]*ast.FuncDecl // Mapping Method Names to Declarations (Multiple Possible per Name)
	methodSet       typeutil.MethodSetCache
	analysisResults map[*ast.FuncDecl]funcAnalysisResult // Mapping Function Declarations to cached error codes and affectors
}

type funcAnalysisResult struct {
	codes     codeSet
	affectors map[ast.Expr]struct{}
}

func newFuncLookup() *funcLookup {
	return &funcLookup{
		map[string]*ast.FuncDecl{},
		map[string][]*ast.FuncDecl{},
		typeutil.MethodSetCache{},
		map[*ast.FuncDecl]funcAnalysisResult{},
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

// searchMethod tries to find the correct function declaration for a method given the receiver type and method name.
func (lookup *funcLookup) searchMethod(pass *analysis.Pass, receiver types.Type, methodName string) *ast.FuncDecl {
	methods, ok := lookup.methods[methodName]
	if !ok || len(methods) == 0 {
		// Return early if there is no method in the current package with the given name
		return nil
	}

	// Search for method in the type information using receiver type and method name
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

func (r *funcAnalysisResult) addAffector(affector ast.Expr) {
	if r.affectors == nil {
		r.affectors = make(map[ast.Expr]struct{})
	}
	r.affectors[affector] = struct{}{}
}

// combineInplace combines the codes and affectors of both funcAnalysisResults.
// The parameter other is modified after this call and should not be used any longer.
func (r *funcAnalysisResult) combineInplace(other funcAnalysisResult) {
	r.codes = unionInplace(r.codes, other.codes)

	// Make sure we add values from the smaller into the bigger set.
	if len(r.affectors) < len(other.affectors) {
		r.affectors, other.affectors = other.affectors, r.affectors
	}

	for value := range other.affectors {
		r.affectors[value] = struct{}{}
	}
}

// combine combines the codes and affectors of both funcAnalysisResults.
// Parameter and reciever of this method both stay unchanged and may be used further.
func (r funcAnalysisResult) combine(other funcAnalysisResult) funcAnalysisResult {
	codes := union(r.codes, other.codes)
	affectors := make(map[ast.Expr]struct{}, len(r.affectors)+len(other.affectors))

	for value := range r.affectors {
		affectors[value] = struct{}{}
	}

	for value := range other.affectors {
		affectors[value] = struct{}{}
	}

	return funcAnalysisResult{codes, affectors}
}
