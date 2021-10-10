package analysis

import (
	"fmt"
	"go/ast"
	"go/types"
	"sort"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

// ErrorInterface is a fact emitted by the analyser,
// marking an interface as containing methods that declare error codes.
type ErrorInterface struct {
	// ErrorMethods contains the names of all methods in the interface,
	// that have error codes declared along with their declared error codes.
	//
	// For all types implementing this interface, these methods must be checked to
	// make sure they only contain a subset of the error codes declared in the interface.
	ErrorMethods map[string]codeSet
}

func (*ErrorInterface) AFact() {}

func (e *ErrorInterface) String() string {
	methods := make([]string, 0, len(e.ErrorMethods))
	for method := range e.ErrorMethods {
		methods = append(methods, method)
	}
	sort.Strings(methods)
	return fmt.Sprintf("ErrorInterface: %v", strings.Join(methods, " "))
}

// errorInterfaceWithCodes is used for temporary storing and passing an interface containing
// methods that declare error codes.
type errorInterfaceWithCodes struct {
	InterfaceIdent *ast.Ident
	ErrorMethods   map[*ast.Ident]codeSet
}

func findErrorReturningInterfaces(pass *analysis.Pass) []*errorInterfaceWithCodes {
	var result []*errorInterfaceWithCodes
	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	// We only need to see type declarations.
	nodeFilter := []ast.Node{
		(*ast.GenDecl)(nil),
	}

	inspect.Nodes(nodeFilter, func(node ast.Node, _ bool) bool {
		genDecl := node.(*ast.GenDecl)

		for _, spec := range genDecl.Specs {
			errorInterface := checkIfErrorReturningInterface(pass, spec)
			if errorInterface != nil && len(errorInterface.ErrorMethods) > 0 {
				result = append(result, errorInterface)
			}
		}

		// Never recurse deeper.
		return false
	})

	return result
}

func checkIfErrorReturningInterface(pass *analysis.Pass, spec ast.Spec) *errorInterfaceWithCodes {
	typeSpec, ok := spec.(*ast.TypeSpec)
	if !ok {
		return nil
	}

	// Make sure type spec is a valid interface.
	interfaceType, ok := typeSpec.Type.(*ast.InterfaceType)
	if !ok || interfaceType.Methods == nil || len(interfaceType.Methods.List) == 0 {
		return nil
	}

	result := errorInterfaceWithCodes{typeSpec.Name, map[*ast.Ident]codeSet{}}

	for _, method := range interfaceType.Methods.List {
		funcType, ok := method.Type.(*ast.FuncType)
		if !ok || !checkFunctionReturnsError(pass, funcType) {
			continue
		}

		methodIdent := method.Names[0]
		codes, err := findErrorDocs(method.Doc)
		if err != nil {
			pass.ReportRangef(method, "interface method %q has odd docstring: %s", methodIdent.Name, err)
			continue
		}

		if len(codes) == 0 {
			// TODO: Exclude Cause() methods of error types from having to declare error codes.

			// Warn directly about any methods if they return errors, but don't declare error codes in their docs.
			pass.ReportRangef(method, "interface method %q does not declare any error codes", methodIdent.Name)
		} else {
			result.ErrorMethods[methodIdent] = codes
		}
	}

	return &result
}

// exportInterfaceFacts exports all codes for each method in each interface as facts,
// additionally exports for each interface the fact that it is an error interface.
func exportInterfaceFacts(pass *analysis.Pass, interfaces []*errorInterfaceWithCodes) {
	for _, errorInterface := range interfaces {
		exportErrorInterfaceFact(pass, errorInterface)
		for methodIdent, codes := range errorInterface.ErrorMethods {
			exportErrorCodesFact(pass, methodIdent, codes)
		}
	}
}

func exportErrorInterfaceFact(pass *analysis.Pass, errorInterface *errorInterfaceWithCodes) {
	interfaceType, ok := pass.TypesInfo.Defs[errorInterface.InterfaceIdent]
	if !ok {
		logf("Could not find definition for interface %q!", errorInterface.InterfaceIdent.Name)
		return
	}

	methods := make(map[string]codeSet, len(errorInterface.ErrorMethods))
	for methodIdent, codes := range errorInterface.ErrorMethods {
		methods[methodIdent.Name] = codes
	}

	fact := ErrorInterface{methods}
	pass.ExportObjectFact(interfaceType, &fact)
}

// findConversionsToErrorReturningInterfaces finds all conversions (implicit or explicit) to
// error returning interfaces. For those conversions we check if the origin type fulfills the
// error code contract of the target interface.
//
// Conversions can happen in many statements and expressions:
// Explicit:
//     - Conversion to Interface
// Implicit:
//     - Assignment
//     - Function Call
//     - Composite Literal:
//         - Struct Creation
//         - Array Creation
//         - Map Creation
//     - Return Statement
func findConversionsToErrorReturningInterfaces(pass *analysis.Pass, lookup *funcLookup) {
	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	var currentFunc *ast.FuncDecl
	var funcLitStack []*ast.FuncLit

	inspect.Nodes(nil, func(node ast.Node, push bool) bool {
		switch node := node.(type) {
		case *ast.FuncDecl:
			if len(funcLitStack) != 0 {
				panic("Should never visit funcDecl when function literals are not completly handled.")
			}

			if push {
				currentFunc = node // enter funcDecl
			} else {
				currentFunc = nil // exit funcDecl
			}
		case *ast.FuncLit:
			if push {
				funcLitStack = append(funcLitStack, node)
			} else {
				if len(funcLitStack) == 0 {
					panic("Cannot remove function literal from stack, because stack was empty.")
				}
				funcLitStack = funcLitStack[:len(funcLitStack)-1]
			}
		}

		if !push {
			return false
		}

		switch node := node.(type) {
		case *ast.ReturnStmt:
			if len(funcLitStack) > 0 {
				findConversionsInReturnStmt(pass, lookup, node, funcLitStack[len(funcLitStack)-1].Type)
			} else if currentFunc != nil {
				findConversionsInReturnStmt(pass, lookup, node, currentFunc.Type)
			} else {
				panic("Found unexpected return statement: returning outside of function or function literal.")
			}
		}

		// Always recurse deeper.
		return true
	})
}

func findConversionsInReturnStmt(pass *analysis.Pass, lookup *funcLookup, statement *ast.ReturnStmt, within *ast.FuncType) {
	if len(statement.Results) == 0 {
		return
	}

	position := 0
	for _, resultField := range within.Results.List {
		nextPosition := position
		if len(resultField.Names) == 0 {
			nextPosition++
		} else {
			nextPosition += len(resultField.Names)
		}

		resultType := pass.TypesInfo.TypeOf(resultField.Type)
		namedType := getNamedType(resultType)

		var errorInterface ErrorInterface
		if namedType != nil && pass.ImportObjectFact(namedType.Obj(), &errorInterface) {
			for i := position; i < nextPosition; i++ {
				expression := statement.Results[i]
				checkIfValidSubtypeForInterface(pass, lookup, &errorInterface, resultType, expression)
			}
		}

		position = nextPosition
	}
}

func checkIfValidSubtypeForInterface(pass *analysis.Pass, lookup *funcLookup, errorInterface *ErrorInterface, interfaceType types.Type, expression ast.Expr) {
	exprType := pass.TypesInfo.TypeOf(expression)
	if types.Identical(exprType, interfaceType) {
		return
	}

	for methodName, interfaceCodes := range errorInterface.ErrorMethods {
		methodType := lookup.searchMethodType(pass, exprType, methodName)
		if methodType == nil {
			panic("Should be unreachable: the given expression was confirmed to implement the interface by the type checker.")
		}

		var implementedCodes ErrorCodes
		if !pass.ImportObjectFact(methodType.Obj(), &implementedCodes) {
			continue
		}

		unexpectedCodes := difference(sliceToSet(implementedCodes.Codes), interfaceCodes)
		if len(unexpectedCodes) > 0 {
			namedType := getNamedType(interfaceType)
			pass.ReportRangef(expression, "cannot use expression as %q value in return statement: method %q declares the following error codes which were not part of the interface: %v", namedType.Obj().Name(), methodName, unexpectedCodes.slice())
		}
	}
}
