package analysis

import (
	"fmt"
	"go/ast"
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
	// that have error codes declared.
	//
	// For all types implementing this interface, these methods must be checked to
	// make sure they only contain a subset of the error codes declared in the interface.
	ErrorMethods []string
}

func (*ErrorInterface) AFact() {}

func (e *ErrorInterface) String() string {
	sort.Strings(e.ErrorMethods)
	return fmt.Sprintf("ErrorInterface: %v", strings.Join(e.ErrorMethods, " "))
}

// errorInterfaceWithCodes is used for temporary storing and passing an interface containing
// methods that declare error codes.
type errorInterfaceWithCodes struct {
	InterfaceIdent *ast.Ident
	ErrorMethods   map[*ast.Ident]codeSet
}

// runVerifyInterfaces runs the analysis for interfaces that contain error returning methods.
func runVerifyInterfaces(pass *analysis.Pass) {
	interfaces := findErrorReturningInterfaces(pass)
	exportInterfaceFacts(pass, interfaces)
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

	methods := make([]string, 0, len(errorInterface.ErrorMethods))
	for methodIdent := range errorInterface.ErrorMethods {
		methods = append(methods, methodIdent.Name)
	}

	fact := ErrorInterface{methods}
	pass.ExportObjectFact(interfaceType, &fact)
}
