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

// errorInterfaceInternal is used for temporary storing and passing an interface containing
// methods that declare error codes.
type errorInterfaceInternal struct {
	interfaceIdent     *ast.Ident
	errorMethods       map[string]*errorMethod
	embeddedInterfaces []ast.Expr
}

type errorMethod struct {
	ident *ast.Ident
	codes funcCodes
}

func findErrorReturningInterfaces(pass *analysis.Pass) []*errorInterfaceInternal {
	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	// We only need to see type declarations.
	nodeFilter := []ast.Node{
		(*ast.GenDecl)(nil),
	}

	errorInterfaces := map[string]*errorInterfaceInternal{}
	embeddingInterfaces := map[string]*errorInterfaceInternal{}

	inspect.Nodes(nodeFilter, func(node ast.Node, _ bool) bool {
		genDecl := node.(*ast.GenDecl)

		for _, spec := range genDecl.Specs {
			errorInterface := checkIfErrorReturningInterface(pass, spec)
			if errorInterface != nil {
				if len(errorInterface.embeddedInterfaces) > 0 {
					embeddingInterfaces[errorInterface.interfaceIdent.Name] = errorInterface
				} else if len(errorInterface.errorMethods) > 0 {
					errorInterfaces[errorInterface.interfaceIdent.Name] = errorInterface
				}
			}
		}

		// Never recurse deeper.
		return false
	})

	findErrorReturningEmbeddingInterfaces(pass, errorInterfaces, embeddingInterfaces)

	var result []*errorInterfaceInternal
	for _, errorInterface := range errorInterfaces {
		result = append(result, errorInterface)
	}
	return result
}

func checkIfErrorReturningInterface(pass *analysis.Pass, spec ast.Spec) *errorInterfaceInternal {
	typeSpec, ok := spec.(*ast.TypeSpec)
	if !ok {
		return nil
	}

	// Make sure type spec is a valid interface.
	interfaceType, ok := typeSpec.Type.(*ast.InterfaceType)
	if !ok || interfaceType.Methods == nil || len(interfaceType.Methods.List) == 0 {
		return nil
	}

	result := errorInterfaceInternal{typeSpec.Name, map[string]*errorMethod{}, nil}

	for _, element := range interfaceType.Methods.List {
		switch elementType := element.Type.(type) {
		case *ast.FuncType: // method declaration
			// Figure out if method returns errors and try to get error code declarations.
			errorMethod, err := checkIfInterfaceMethodDeclaresErrors(pass, interfaceType, element, elementType)
			if err != nil {
				pass.ReportRangef(element, "%v", err)
			} else if errorMethod != nil {
				result.errorMethods[errorMethod.ident.Name] = errorMethod
			}
		case *ast.Ident, *ast.SelectorExpr: // embedded interface
			// Remember idents of embedded interfaces for later processing.
			result.embeddedInterfaces = append(result.embeddedInterfaces, elementType)
		}
	}

	return &result
}

func checkIfInterfaceMethodDeclaresErrors(pass *analysis.Pass, interfaceType *ast.InterfaceType, method *ast.Field, funcType *ast.FuncType) (*errorMethod, error) {
	if !checkFunctionReturnsError(pass, funcType) {
		return nil, nil
	}

	methodIdent := method.Names[0]
	codes, errorCodeParamName, declaredNoCodesOk, err := findErrorDocs(method.Doc)
	if err != nil {
		return nil, fmt.Errorf("interface method %q has odd docstring: %s", methodIdent.Name, err)
	}

	// TODO: Implement support, then remove this check
	if errorCodeParamName != "" {
		return nil, fmt.Errorf("declaration of error constructors in interfaces is currently not supported")
	}

	errorCodeParam, ok := findErrorCodeParamIdent(pass, funcType, errorCodeParamName)
	if !ok {
		return nil, nil
	}

	if len(codes) == 0 && !declaredNoCodesOk && errorCodeParam == nil {
		// Exclude Cause() methods of error types from having to declare error codes.
		interfaceType := pass.TypesInfo.TypeOf(interfaceType)
		if methodIdent.Name == "Cause" && types.Implements(interfaceType, tReeErrorWithCause) {
			return nil, nil
		}

		// Warn directly about any methods if they return errors, but don't declare error codes in their docs.
		return nil, fmt.Errorf("interface method %q does not declare any error codes", methodIdent.Name)
	} else {
		return &errorMethod{methodIdent, funcCodes{codes, errorCodeParam}}, nil
	}
}

// findErrorReturningEmbeddingInterfaces searches through interfaces that embedd other interfaces starting with the given interface.
//
// Each visited interface is checked if it contains error returning methods.
// If so, those methods are checked against methods with the equal name in other embedded interfaces (and methods on the current interface).
// Diagnostics are emitted if declared error codes of equally named methods do not match.
func findErrorReturningEmbeddingInterfaces(pass *analysis.Pass, errorInterfaces map[string]*errorInterfaceInternal, embeddingInterfaces map[string]*errorInterfaceInternal) {
	embeddedInterfaceNames := make([]string, 0, len(embeddingInterfaces))
	for name := range embeddingInterfaces {
		embeddedInterfaceNames = append(embeddedInterfaceNames, name)
	}

	for _, embeddedName := range embeddedInterfaceNames {
		embedded, ok := embeddingInterfaces[embeddedName]
		if ok { // Visit the embedded interface if it was not yet visited.
			embeddingInterfaceDFS(pass, errorInterfaces, embeddingInterfaces, embedded)
		}
	}
}

func embeddingInterfaceDFS(pass *analysis.Pass, errorInterfaces map[string]*errorInterfaceInternal, embeddingInterfaces map[string]*errorInterfaceInternal, embedding *errorInterfaceInternal) {
	// Mark given interface as visited.
	delete(embeddingInterfaces, embedding.interfaceIdent.Name)

	for _, embedded := range embedding.embeddedInterfaces {
		embeddedIdent, ok := embedded.(*ast.Ident)
		if !ok {
			exprType := pass.TypesInfo.TypeOf(embedded)
			errorInterface := importErrorInterfaceFact(pass, exprType)
			if errorInterface != nil {
				addEmbeddedInterfaceErrorMethodsForFact(pass, embedding, errorInterface, embedded)
			}
			continue
		}

		// Handle embedded interfaces first.
		// If they contain methods that declare errors, we can add those errors to the current interface too.
		embeddedEmbedding, ok := embeddingInterfaces[embeddedIdent.Name]
		if ok {
			embeddingInterfaceDFS(pass, errorInterfaces, embeddingInterfaces, embeddedEmbedding)
		}

		errorInterface, ok := errorInterfaces[embeddedIdent.Name]
		if ok {
			addEmbeddedInterfaceErrorMethods(pass, embedding, errorInterface, embedded)
		}
	}

	if len(embedding.errorMethods) > 0 {
		errorInterfaces[embedding.interfaceIdent.Name] = embedding
	}
}

// addEmbeddedInterfaceErrorMethods adds all error methods from one given interface (add) to the other given interface (embedding).
//
// If an error method already exists in the target interface, the defined error codes are compared and
// diagnostics are emitted if they don't match.
func addEmbeddedInterfaceErrorMethods(pass *analysis.Pass, embedding *errorInterfaceInternal, add *errorInterfaceInternal, reportPos analysis.Range) {
	for methodName, newErrorMethod := range add.errorMethods {
		oldErrorMethod, ok := embedding.errorMethods[methodName]
		if !ok {
			embedding.errorMethods[methodName] = &errorMethod{nil, newErrorMethod.codes}
			continue
		}

		checkEmbeddedInterfaceErrorMethodCodes(pass, oldErrorMethod.codes.codes, newErrorMethod.codes.codes, methodName, reportPos)
	}
}

func addEmbeddedInterfaceErrorMethodsForFact(pass *analysis.Pass, embedding *errorInterfaceInternal, add *ErrorInterface, reportPos analysis.Range) {
	for methodName, newErrorMethodCodes := range add.ErrorMethods {
		oldErrorMethod, ok := embedding.errorMethods[methodName]
		if !ok {
			embedding.errorMethods[methodName] = &errorMethod{nil, funcCodes{newErrorMethodCodes, nil}}
			continue
		}

		checkEmbeddedInterfaceErrorMethodCodes(pass, oldErrorMethod.codes.codes, newErrorMethodCodes, methodName, reportPos)
	}
}

// checkEmbeddedInterfaceErrorMethodCodes checks if new and old codes of methods with equal name are compatible.
func checkEmbeddedInterfaceErrorMethodCodes(pass *analysis.Pass, oldCodes CodeSet, newCodes CodeSet, methodName string, reportPos analysis.Range) {
	errorCodesMatch, errorMessage := checkIfErrorCodesMatch(oldCodes, newCodes)
	if !errorCodesMatch {
		pass.ReportRangef(reportPos, "embedded interface is not compatible: method %q has mismatches in declared error codes: %s", methodName, errorMessage)
	}
}

// exportInterfaceFacts exports all codes for each method in each interface as facts,
// additionally exports for each interface the fact that it is an error interface.
func exportInterfaceFacts(pass *analysis.Pass, interfaces []*errorInterfaceInternal) {
	for _, errorInterface := range interfaces {
		exportErrorInterfaceFact(pass, errorInterface)
		for _, errorMethod := range errorInterface.errorMethods {
			if errorMethod.ident == nil {
				continue
			}

			if errorMethod.codes.param != nil {
				exportErrorConstructorFact(pass, errorMethod.ident, errorMethod.codes.param)
			}
			exportErrorCodesFact(pass, errorMethod.ident, errorMethod.codes.codes)
		}
	}
}

func exportErrorInterfaceFact(pass *analysis.Pass, errorInterface *errorInterfaceInternal) {
	interfaceType, ok := pass.TypesInfo.Defs[errorInterface.interfaceIdent]
	if !ok {
		logf("Could not find definition for interface %q!", errorInterface.interfaceIdent.Name)
		return
	}

	methods := make(map[string]CodeSet, len(errorInterface.errorMethods))
	for methodName, errorMethod := range errorInterface.errorMethods {
		methods[methodName] = errorMethod.codes.codes
	}

	fact := ErrorInterface{methods}
	pass.ExportObjectFact(interfaceType, &fact)
}
