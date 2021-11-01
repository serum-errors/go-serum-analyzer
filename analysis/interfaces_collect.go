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
	ErrorMethods map[string]CodeSet
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
	interfaceIdent *ast.Ident
	errorMethods   map[*ast.Ident]funcCodes
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
			if errorInterface != nil && len(errorInterface.errorMethods) > 0 {
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

	result := errorInterfaceWithCodes{typeSpec.Name, map[*ast.Ident]funcCodes{}}

	for _, method := range interfaceType.Methods.List {
		funcType, ok := method.Type.(*ast.FuncType)
		if !ok || !checkFunctionReturnsError(pass, funcType) {
			continue
		}

		methodIdent := method.Names[0]
		codes, errorCodeParamName, declaredNoCodesOk, err := findErrorDocs(method.Doc)
		if err != nil {
			pass.ReportRangef(method, "interface method %q has odd docstring: %s", methodIdent.Name, err)
			continue
		}

		// TODO: Implement support, then remove this check
		if errorCodeParamName != "" {
			pass.ReportRangef(method, "declaration of error constructors in interfaces is currently not supported")
			continue
		}

		errorCodeParam, ok := findErrorCodeParamIdent(pass, funcType, errorCodeParamName)
		if !ok {
			continue
		}

		if len(codes) == 0 && !declaredNoCodesOk && errorCodeParam == nil {
			// Exclude Cause() methods of error types from having to declare error codes.
			interfaceType := pass.TypesInfo.TypeOf(typeSpec.Type)
			if methodIdent.Name == "Cause" && types.Implements(interfaceType, tReeErrorWithCause) {
				continue
			}

			// Warn directly about any methods if they return errors, but don't declare error codes in their docs.
			pass.ReportRangef(method, "interface method %q does not declare any error codes", methodIdent.Name)
		} else {
			result.errorMethods[methodIdent] = funcCodes{codes, errorCodeParam}
		}
	}

	return &result
}

// exportInterfaceFacts exports all codes for each method in each interface as facts,
// additionally exports for each interface the fact that it is an error interface.
func exportInterfaceFacts(pass *analysis.Pass, interfaces []*errorInterfaceWithCodes) {
	for _, errorInterface := range interfaces {
		exportErrorInterfaceFact(pass, errorInterface)
		for methodIdent, funcCodes := range errorInterface.errorMethods {
			if funcCodes.param != nil {
				exportErrorConstructorFact(pass, methodIdent, funcCodes.param)
			}
			exportErrorCodesFact(pass, methodIdent, funcCodes.codes)
		}
	}
}

func exportErrorInterfaceFact(pass *analysis.Pass, errorInterface *errorInterfaceWithCodes) {
	interfaceType, ok := pass.TypesInfo.Defs[errorInterface.interfaceIdent]
	if !ok {
		logf("Could not find definition for interface %q!", errorInterface.interfaceIdent.Name)
		return
	}

	methods := make(map[string]CodeSet, len(errorInterface.errorMethods))
	for methodIdent, funcCodes := range errorInterface.errorMethods {
		methods[methodIdent.Name] = funcCodes.codes
	}

	fact := ErrorInterface{methods}
	pass.ExportObjectFact(interfaceType, &fact)
}
