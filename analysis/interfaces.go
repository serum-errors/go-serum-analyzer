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
	InterfaceIdent *ast.Ident
	ErrorMethods   map[*ast.Ident]CodeSet
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

	result := errorInterfaceWithCodes{typeSpec.Name, map[*ast.Ident]CodeSet{}}

	for _, method := range interfaceType.Methods.List {
		funcType, ok := method.Type.(*ast.FuncType)
		if !ok || !checkFunctionReturnsError(pass, funcType) {
			continue
		}

		methodIdent := method.Names[0]
		codes, declaredNoCodesOk, err := findErrorDocs(method.Doc)
		if err != nil {
			pass.ReportRangef(method, "interface method %q has odd docstring: %s", methodIdent.Name, err)
			continue
		}

		if len(codes) == 0 && !declaredNoCodesOk {
			// Exclude Cause() methods of error types from having to declare error codes.
			interfaceType := pass.TypesInfo.TypeOf(typeSpec.Type)
			if methodIdent.Name == "Cause" && types.Implements(interfaceType, tReeErrorWithCause) {
				continue
			}

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

	methods := make(map[string]CodeSet, len(errorInterface.ErrorMethods))
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
//     - Type Assertion
// Implicit:
//     - Assignment
//     - Function Call
//     - Composite Literal:
//         - Struct Creation
//         - Array Creation
//         - Slice Creation
//         - Map Creation
//     - Return Statement
//     - Map Index
//     - Range Statement
//     - Channel Send
func findConversionsToErrorReturningInterfaces(c *context) {
	inspect := c.pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	var currentFunc *ast.FuncDecl
	var funcLitStack []*ast.FuncLit

	inspect.Nodes(nil, func(node ast.Node, push bool) bool {
		switch node := node.(type) {
		case *ast.FuncDecl:
			if len(funcLitStack) != 0 {
				panic("should never visit funcDecl when function literals are not completly handled.")
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
					panic("cannot remove function literal from stack, because stack was empty.")
				}
				funcLitStack = funcLitStack[:len(funcLitStack)-1]
			}
		}

		if !push {
			return false
		}

		switch node := node.(type) {
		case *ast.AssignStmt:
			findConversionsInAssignStmt(c, node)
		case *ast.ValueSpec:
			findConversionsInValueSpec(c, node)
		case *ast.CallExpr:
			findConversionsInCallExpr(c, node)
		case *ast.TypeAssertExpr:
			findConversionsInTypeAssertExpr(c, node)
		case *ast.TypeSwitchStmt:
			findConversionsInTypeSwitchStmt(c, node)
		case *ast.IndexExpr:
			findConversionsInIndexExpr(c, node)
		case *ast.CompositeLit:
			findConversionsInCompositeLit(c, node)
		case *ast.ReturnStmt:
			if len(funcLitStack) > 0 {
				findConversionsInReturnStmt(c, node, funcLitStack[len(funcLitStack)-1].Type)
			} else if currentFunc != nil {
				findConversionsInReturnStmt(c, node, currentFunc.Type)
			} else {
				panic("found unexpected return statement: returning outside of function or function literal.")
			}
		case *ast.RangeStmt:
			findConversionsInRangeStmtKey(c, node)
			findConversionsInRangeStmtValue(c, node)
		case *ast.SendStmt:
			findConversionsInSendStmt(c, node)
		}

		// Always recurse deeper.
		return true
	})
}

func findConversionsInAssignStmt(c *context, statement *ast.AssignStmt) {
	pass := c.pass

	for i, lhsEntry := range statement.Lhs {
		lhsType := pass.TypesInfo.TypeOf(lhsEntry)
		errorInterface := importErrorInterfaceFact(pass, lhsType)
		if errorInterface == nil {
			continue
		}

		if len(statement.Lhs) == len(statement.Rhs) { // Rhs is comma separated
			expression := statement.Rhs[i]
			checkIfExprHasValidSubtypeForInterface(c, errorInterface, lhsType, expression)
		} else { // Rhs is a function call
			callExpr := statement.Rhs[0]
			callType, ok := pass.TypesInfo.TypeOf(callExpr).(*types.Tuple)
			if !ok || i >= callType.Len() {
				panic("should be unreachable: function call destructuring should always be of type tuple with sufficient length")
			}

			exprType := callType.At(i).Type()
			checkIfTypeIsValidSubtypeForInterface(c, errorInterface, lhsType, exprType, callExpr)
		}
	}
}

func findConversionsInValueSpec(c *context, spec *ast.ValueSpec) {
	if len(spec.Values) == 0 || spec.Type == nil {
		return
	}

	pass := c.pass
	specType := pass.TypesInfo.TypeOf(spec.Type)
	errorInterface := importErrorInterfaceFact(pass, specType)
	if errorInterface == nil {
		return
	}

	if len(spec.Names) == len(spec.Values) { // right hand side is comma separated
		for _, value := range spec.Values {
			checkIfExprHasValidSubtypeForInterface(c, errorInterface, specType, value)
		}
	} else { // right hand side is a function call
		callExpr := spec.Values[0]
		callType, ok := pass.TypesInfo.TypeOf(callExpr).(*types.Tuple)
		if !ok {
			panic("should be unreachable: function call destructuring should always be of type tuple with sufficient length")
		}

		for i := range spec.Names {
			exprType := callType.At(i).Type()
			checkIfTypeIsValidSubtypeForInterface(c, errorInterface, specType, exprType, callExpr)
		}
	}
}

func findConversionsInCallExpr(c *context, callExpr *ast.CallExpr) {
	if len(callExpr.Args) == 0 {
		return
	}

	pass := c.pass
	functionType := pass.TypesInfo.TypeOf(callExpr.Fun)
	signature, ok := functionType.(*types.Signature)
	if !ok {
		// The given call expression is a type conversion.
		findConversionsExplicit(c, callExpr, functionType)
	} else {
		// The given call expression is a regular call to a function.
		for i := 0; i < signature.Params().Len(); i++ {
			paramType := signature.Params().At(i).Type()
			errorInterface := importErrorInterfaceFact(pass, paramType)
			if errorInterface == nil {
				continue
			}

			checkIfExprHasValidSubtypeForInterface(c, errorInterface, paramType, callExpr.Args[i])

			if signature.Variadic() && i == signature.Params().Len()-1 {
				for j := signature.Params().Len(); j < len(callExpr.Args); j++ {
					checkIfExprHasValidSubtypeForInterface(c, errorInterface, paramType, callExpr.Args[j])
				}
			}
		}

		// Handle variadic parameters at the end of argument list.
		if signature.Variadic() {
			paramType := signature.Params().At(signature.Params().Len() - 1).Type()
			sliceType, ok := paramType.(*types.Slice)
			if !ok {
				// This is the case for some append function calls with string type...
				return
			}

			errorInterface := importErrorInterfaceFact(pass, sliceType.Elem())
			if errorInterface == nil {
				return
			}

			for i := signature.Params().Len() - 1; i < len(callExpr.Args); i++ {
				checkIfExprHasValidSubtypeForInterface(c, errorInterface, sliceType.Elem(), callExpr.Args[i])
			}
		}
	}
}

func findConversionsExplicit(c *context, callExpr *ast.CallExpr, targetType types.Type) {
	errorInterface := importErrorInterfaceFact(c.pass, targetType)
	if errorInterface == nil {
		return
	}

	if len(callExpr.Args) != 1 {
		panic("should be unreachable: type conversion may only have one parameter")
	}

	checkIfExprHasValidSubtypeForInterface(c, errorInterface, targetType, callExpr.Args[0])
}

func findConversionsInTypeAssertExpr(c *context, typeAssertExpr *ast.TypeAssertExpr) {
	if typeAssertExpr.Type == nil {
		return // Ignore all "switch X.(type) { ... }" kind of type assertions.
	}

	pass := c.pass
	targetType := pass.TypesInfo.TypeOf(typeAssertExpr.Type)
	errorInterface := importErrorInterfaceFact(pass, targetType)
	if errorInterface == nil {
		return
	}

	checkIfExprHasValidSubtypeForInterface(c, errorInterface, targetType, typeAssertExpr.X)
}

func findConversionsInTypeSwitchStmt(c *context, typeSwitchStmt *ast.TypeSwitchStmt) {
	if typeSwitchStmt.Body == nil {
		return
	}

	// Only consider cases where an invalid type might be assigned to a variable.
	assignment, ok := typeSwitchStmt.Assign.(*ast.AssignStmt)
	if !ok || len(assignment.Rhs) != 1 {
		return
	}
	expression := assignment.Rhs[0].(*ast.TypeAssertExpr)
	exprType := c.pass.TypesInfo.TypeOf(expression.X)

	pass := c.pass

	for _, caseClause := range typeSwitchStmt.Body.List {
		caseClause := caseClause.(*ast.CaseClause)
		for _, caseElement := range caseClause.List {
			caseType := pass.TypesInfo.TypeOf(caseElement)
			errorInterface := importErrorInterfaceFact(pass, caseType)
			if errorInterface == nil {
				continue
			}

			namedType := getNamedType(caseType)
			if namedType == nil {
				continue
			}

			// Check that the switch expression actually implements the interface in the case.
			caseTypeInterface, ok := namedType.Underlying().(*types.Interface)
			if !ok || !types.Implements(exprType, caseTypeInterface) {
				continue
			}

			checkIfTypeIsValidSubtypeForInterface(c, errorInterface, caseType, exprType, caseElement)
		}
	}
}

func findConversionsInIndexExpr(c *context, indexExpr *ast.IndexExpr) {
	pass := c.pass
	mapType, ok := pass.TypesInfo.TypeOf(indexExpr.X).(*types.Map)
	if !ok {
		// indexExpr.X can be a map or any of slice, array, pointer to array, or string.
		// In case it is not a map, the index will be an integer and not relevant for the following checks.
		return
	}

	errorInterface := importErrorInterfaceFact(pass, mapType.Key())
	if errorInterface == nil {
		return
	}

	checkIfExprHasValidSubtypeForInterface(c, errorInterface, mapType.Key(), indexExpr.Index)
}

func findConversionsInCompositeLit(c *context, composite *ast.CompositeLit) {
	if len(composite.Elts) == 0 {
		return
	}

	exprType := c.pass.TypesInfo.TypeOf(composite)

	// Unpack named type to find actual type.
	if namedType, ok := exprType.(*types.Named); ok {
		exprType = namedType.Underlying()
	}

	switch exprType := exprType.(type) {
	case *types.Struct:
		findConversionsInStructLit(c, composite, exprType)
	case *types.Slice:
		findConversionsInCompositeValues(c, composite, exprType.Elem())
	case *types.Array:
		findConversionsInCompositeValues(c, composite, exprType.Elem())
	case *types.Map:
		findConversionsInCompositeValues(c, composite, exprType.Elem())
		findConversionsInMapLitKeys(c, composite, exprType.Key())
	default:
		logf("composite lit type not handled in findConversionsInCompositeLit: %#v\n", exprType)
	}
}

func findConversionsInStructLit(c *context, composite *ast.CompositeLit, structType *types.Struct) {
	if len(composite.Elts) == 0 {
		return
	}

	for i := 0; i < structType.NumFields(); i++ {
		field := structType.Field(i)
		fieldType := field.Type()
		errorInterface := importErrorInterfaceFact(c.pass, fieldType)
		if errorInterface == nil {
			continue
		}

		if _, ok := composite.Elts[0].(*ast.KeyValueExpr); !ok { // struct creation has positional arguments
			checkIfExprHasValidSubtypeForInterface(c, errorInterface, fieldType, composite.Elts[i])
		} else { // struct creation has keyed arguments
			for _, expr := range composite.Elts {
				exprKeyed := expr.(*ast.KeyValueExpr) // if one element is key-value, all have to be
				key := exprKeyed.Key.(*ast.Ident)
				if key.Name == field.Name() {
					checkIfExprHasValidSubtypeForInterface(c, errorInterface, fieldType, exprKeyed.Value)
					break
				}
			}
		}
	}
}

func findConversionsInCompositeValues(c *context, composite *ast.CompositeLit, elemType types.Type) {
	errorInterface := importErrorInterfaceFact(c.pass, elemType)
	if errorInterface == nil {
		return
	}

	for _, element := range composite.Elts {
		if keyedElement, ok := element.(*ast.KeyValueExpr); ok {
			element = keyedElement.Value // key is not relevant for the following check
		}
		checkIfExprHasValidSubtypeForInterface(c, errorInterface, elemType, element)
	}
}

func findConversionsInMapLitKeys(c *context, composite *ast.CompositeLit, keyType types.Type) {
	errorInterface := importErrorInterfaceFact(c.pass, keyType)
	if errorInterface == nil {
		return
	}

	for _, element := range composite.Elts {
		keyedElement := element.(*ast.KeyValueExpr) // all elements have to be key-value, because it's a map
		checkIfExprHasValidSubtypeForInterface(c, errorInterface, keyType, keyedElement.Key)
	}
}

func findConversionsInReturnStmt(c *context, statement *ast.ReturnStmt, within *ast.FuncType) {
	if len(statement.Results) == 0 {
		return
	}

	pass := c.pass
	position := 0
	for _, resultField := range within.Results.List {
		nextPosition := position
		if len(resultField.Names) == 0 {
			nextPosition++
		} else {
			nextPosition += len(resultField.Names)
		}

		resultType := pass.TypesInfo.TypeOf(resultField.Type)
		errorInterface := importErrorInterfaceFact(pass, resultType)
		if errorInterface != nil {
			for i := position; i < nextPosition; i++ {
				expression := statement.Results[i]
				checkIfExprHasValidSubtypeForInterface(c, errorInterface, resultType, expression)
			}
		}

		position = nextPosition
	}
}

func findConversionsInRangeStmtKey(c *context, statement *ast.RangeStmt) {
	if statement.Key == nil {
		return
	}

	pass := c.pass
	keyType := pass.TypesInfo.TypeOf(statement.Key)
	errorInterface := importErrorInterfaceFact(pass, keyType)
	if errorInterface == nil {
		return
	}

	var exprType types.Type
	switch rhsType := pass.TypesInfo.TypeOf(statement.X).(type) { // has to be: map or channel
	case *types.Map:
		exprType = rhsType.Key()
	case *types.Chan:
		exprType = rhsType.Elem()
	default:
		logf("unexpected type: %#v\n", rhsType)
		panic("unexpected type in for-range statement")
	}

	checkIfTypeIsValidSubtypeForInterface(c, errorInterface, keyType, exprType, statement.X)
}

func findConversionsInRangeStmtValue(c *context, statement *ast.RangeStmt) {
	if statement.Key == nil {
		return
	}

	pass := c.pass
	valueType := pass.TypesInfo.TypeOf(statement.Value)
	errorInterface := importErrorInterfaceFact(pass, valueType)
	if errorInterface == nil {
		return
	}

	var exprType types.Type
	switch rhsType := pass.TypesInfo.TypeOf(statement.X).(type) { // has to be: pointer to array, array, slice, or map
	case *types.Pointer:
		arrayType := rhsType.Elem().(*types.Array)
		exprType = arrayType.Elem()
	default:
		exprType = rhsType.(interface{ Elem() types.Type }).Elem()
	}

	checkIfTypeIsValidSubtypeForInterface(c, errorInterface, valueType, exprType, statement.X)
}

func findConversionsInSendStmt(c *context, statement *ast.SendStmt) {
	pass := c.pass
	chanType := pass.TypesInfo.TypeOf(statement.Chan).(*types.Chan)
	errorInterface := importErrorInterfaceFact(pass, chanType.Elem())
	if errorInterface == nil {
		return
	}

	checkIfExprHasValidSubtypeForInterface(c, errorInterface, chanType.Elem(), statement.Value)
}

// importErrorInterfaceFact imports and returns the ErrorInterface fact for the given type,
// or returns nil if no such fact exists.
func importErrorInterfaceFact(pass *analysis.Pass, interfaceType types.Type) *ErrorInterface {
	result := new(ErrorInterface)
	namedType := getNamedType(interfaceType)
	if namedType != nil && pass.ImportObjectFact(namedType.Obj(), result) {
		return result
	}
	return nil
}

func checkIfExprHasValidSubtypeForInterface(c *context, errorInterface *ErrorInterface, interfaceType types.Type, expression ast.Expr) {
	exprType := c.pass.TypesInfo.TypeOf(expression)
	checkIfTypeIsValidSubtypeForInterface(c, errorInterface, interfaceType, exprType, expression)
}

// checkIfTypeIsValidSubtypeForInterface checks if the type (exprType) is a valid subtype of the interface type (interfaceType)
// by looking at the error codes returned by the methods of both types.
//
// To be a valid subtype, the error codes of all methods in the implementation must be
// a subset of the error codes of the respective method in the interface.
func checkIfTypeIsValidSubtypeForInterface(c *context, errorInterface *ErrorInterface, interfaceType types.Type, exprType types.Type, exprPos analysis.Range) {
	// If the types are identical, then declared error codes are also identical.
	if types.Identical(exprType, interfaceType) {
		return
	}

	pass, lookup := c.pass, c.lookup

	// Nil values are always ok.
	basicType, ok := exprType.(*types.Basic)
	if ok && basicType.Kind() == types.UntypedNil {
		return
	}

	for methodName, interfaceCodes := range errorInterface.ErrorMethods {
		methodType := lookup.searchMethodType(pass, exprType, methodName)
		if methodType == nil {
			panic("should be unreachable: the given expression was confirmed to implement the interface by the type checker.")
		}

		var foundCodes CodeSet
		var implementedCodes ErrorCodes
		// Try to get error codes from fact.
		if pass.ImportObjectFact(methodType.Obj(), &implementedCodes) {
			foundCodes = implementedCodes.Codes
		} else {
			// Failed: Could be a non-exported function.
			var ok bool
			methodDecl := lookup.searchMethod(pass, exprType, methodName)
			foundCodes, ok = lookup.foundCodes[methodDecl]
			if !ok {
				foundCodes = findErrorCodesInFunc(c, &funcDefinition{methodDecl, nil})
			}
		}

		unexpectedCodes := Difference(foundCodes, interfaceCodes)
		if len(unexpectedCodes) > 0 {
			namedType := getNamedType(interfaceType)
			unexpectedCodes := unexpectedCodes.Slice()
			sort.Strings(unexpectedCodes)
			pass.ReportRangef(exprPos, "cannot use expression as %q value: method %q declares the following error codes which were not part of the interface: %v", namedType.Obj().Name(), methodName, unexpectedCodes)
		}
	}
}
