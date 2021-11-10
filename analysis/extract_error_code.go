package analysis

import (
	"fmt"
	"go/ast"
	"go/constant"
	"go/token"
	"go/types"
	"strconv"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/ast/astutil"
)

// extractErrorCodesFromAffector extracts all error codes from the given affectors and returns them.
func extractErrorCodesFromAffector(pass *analysis.Pass, lookup *funcLookup, function *funcDefinition, affector ast.Expr) CodeSet {
	result := Set()

	// Make sure method "Code() string" is present
	if !checkErrorTypeHasLegibleCode(pass, affector) {
		pass.ReportRangef(affector, "expression does not define an error code")
		return result
	}

	errorType, err := getErrorTypeForError(pass, pass.TypesInfo.Types[affector].Type)
	if err != nil || errorType == nil {
		pass.ReportRangef(affector, "expression is not a valid error: error types must return constant error codes or a single field")
	}
	if err != nil {
		logf("Error while looking at affector: %v (Affector: %#v)\n", err, affector)
	} else if errorType != nil {
		if len(errorType.Codes) > 0 {
			result = Union(result, SliceToSet(errorType.Codes))
		}

		if errorType.Field != nil {
			code, ok := extractFieldErrorCode(pass, affector, function, errorType)
			if ok {
				result.Add(code)
			}
		}
	}

	return result
}

// extractFieldErrorCode finds a possible error code from the given constructor expression.
//
// The expression evaluates to an error of the given error type, which has its errorType.Field set to a value (not nil).
func extractFieldErrorCode(pass *analysis.Pass, expr ast.Expr, function *funcDefinition, errorType *ErrorType) (string, bool) {
	if errorType == nil || errorType.Field == nil {
		panic("cannot extract field error code without field definition")
	}

	fieldExpr := findFieldInitExpression(pass, expr, errorType.Field)
	if fieldExpr == nil {
		return "", false
	}

	return extractErrorCodeFromStringExpression(pass, function, fieldExpr)
}

func findFieldInitExpression(pass *analysis.Pass, constructExpr ast.Expr, field *ErrorCodeField) ast.Expr {
	switch expr := astutil.Unparen(constructExpr).(type) {
	case *ast.CompositeLit:
		if len(expr.Elts) == 0 {
			// If no elements are present, the code is being initialised to empty string.
			// This is always a valid value, but never reported.
			return nil
		}

		// Key-based composite literal:
		// Use the field name to find the error code.
		var isKeyBased bool
		for _, element := range expr.Elts {
			element, ok := element.(*ast.KeyValueExpr)
			if !ok { // Either all elements are KeyValueExpr or none.
				break
			}
			isKeyBased = true

			ident, ok := element.Key.(*ast.Ident)
			if !ok {
				logf("found weird key %#v in composite literal %#v\n", element.Key, expr)
				break
			}

			if field.Name == ident.Name {
				return element.Value
			}
		}

		if isKeyBased {
			// If the key is not present, the code is being initialised to empty string.
			return nil
		}

		// Position-based composite literal:
		// Use the field position to find the error code.
		pos := field.Position
		if pos < len(expr.Elts) {
			return expr.Elts[pos]
		}
	case *ast.UnaryExpr:
		if expr.Op == token.AND {
			return findFieldInitExpression(pass, expr.X, field)
		}
	default:
		logf("findFieldInitExpression did not yet handle: %#v\n", expr)
	}

	pass.ReportRangef(constructExpr, "could not find initialiser for error code field in contructor expression")
	return nil
}

func extractErrorCodeFromConstructorCall(pass *analysis.Pass, startingFunc *funcDefinition, reportRange analysis.Range, callee types.Object, callExpr *ast.CallExpr) (string, bool) {
	var fact ErrorConstructor
	if callee == nil || !pass.ImportObjectFact(callee, &fact) {
		return "", false
	}

	if callExpr == nil {
		pass.ReportRangef(reportRange, "unsupported use of error constructor %q", callee.Name())
		return "", false
	}

	if fact.CodeParamPosition >= len(callExpr.Args) {
		panic("should be unreachable: found function call using less arguments than defined in the function's parameter list")
	}

	return extractErrorCodeFromStringExpression(pass, startingFunc, callExpr.Args[fact.CodeParamPosition])
}

func extractErrorCodeFromStringExpression(pass *analysis.Pass, function *funcDefinition, codeExpr ast.Expr) (string, bool) {
	info, ok := pass.TypesInfo.Types[codeExpr]
	if ok && info.Value != nil {
		code, err := getErrorCodeFromConstant(info.Value)
		if err != nil {
			pass.ReportRangef(codeExpr, "%v", err)
		}
		return code, err == nil && code != ""
	}

	// function might be an error constructor and codeExpr the error code parameter.
	fieldExprIdent, ok := astutil.Unparen(codeExpr).(*ast.Ident)
	paramPosition := -1
	if ok {
		paramPosition = getParamPosition(function.Type(), fieldExprIdent)
	}

	if paramPosition >= 0 {
		checkIfExprIsErrorCodeParam(pass, function, &funcCodeParam{fieldExprIdent, paramPosition})
	} else {
		pass.ReportRangef(codeExpr, "error code has to be constant value or error code parameter")
	}

	return "", false
}

func getErrorCodeFromConstant(value constant.Value) (string, error) {
	if value.Kind() != constant.String {
		// Should not be reachable, because we already checked the signature of Code() to return a string.
		// And the value is in the end one that gets returned by Code().
		// So there should be a compiler error if value is not of type string.
		return "", fmt.Errorf("error code has to be of type string")
	}

	result := value.String()
	result, err := strconv.Unquote(result)
	if err != nil {
		return "", fmt.Errorf("problem unquoting string constant value: %v", err)
	}

	if result != "" && !isErrorCodeValid(result) {
		return "", fmt.Errorf("error code has invalid format: should match [a-zA-Z][a-zA-Z0-9\\-]*[a-zA-Z0-9]")
	}

	return result, nil
}

// checkIfExprIsErrorCodeParam checks if the given function is an error constructor and
// the error code parameter is at the given position in the function definition.
//
// If none of these are the case, diagnostics are emitted.
func checkIfExprIsErrorCodeParam(pass *analysis.Pass, function *funcDefinition, param *funcCodeParam) {
	ok := func() bool {
		var fact ErrorConstructor
		if !importErrorConstructorFact(pass, function, &fact) {
			return false
		}
		return param.position == fact.CodeParamPosition
	}()

	if !ok {
		pass.ReportRangef(param.ident, "require an error code parameter declaration to use %q as an error code", param.ident.Name)
	}
}

// ectractErrorCodesFromConstructor checks if the given function is an error constructor
// and if the error code parameter is used correctly.
//
// ectractErrorCodesFromConstructor returns all error codes that were assigned to the error code parameter.
//
// Error code parameters may only be assigned constant strings and local variables and
// pointers to them or assigned local variable may not be leaked.
func ectractErrorCodesFromConstructor(c *context, function *funcDefinition) CodeSet {
	pass := c.pass
	result := Set()

	var fact ErrorConstructor
	if !importErrorConstructorFact(pass, function, &fact) {
		return result
	}

	paramIdent := getParamIdent(function.Type(), fact.CodeParamPosition)
	if paramIdent == nil {
		logf("%v invalid fact: %v\n", pass.Fset.Position(function.node().Pos()), fact)
		panic("should be unreachable: error constructor fact points to parameter that does not exist")
	}

	taintResult := taintSpreadForParamIdentOfImmutableType(pass, paramIdent, function)

	for _, badIdent := range taintResult.identOutOfScope {
		pass.ReportRangef(badIdent, "error code parameter may not be assigned an other parameter, receiver or global variable")
	}

	for _, destruct := range taintResult.destructAssignment {
		pass.ReportRangef(destruct.source, "unsupported: assigning result of function call to error code parameter %q is not allowed", destruct.target.Name)
	}

	for _, expr := range taintResult.expressions {
		code, ok := extractErrorCodeFromStringExpression(pass, function, expr)
		if ok {
			result.Add(code)
		}
	}

	return result
}

// importErrorConstructorFact tries to import the ErrorConstructor fact for the given function.
func importErrorConstructorFact(pass *analysis.Pass, function *funcDefinition, fact *ErrorConstructor) bool {
	if function == nil || function.funcDecl == nil {
		return false
	}

	funcName := function.funcDecl.Name
	funcObj := pass.TypesInfo.ObjectOf(funcName)
	return pass.ImportObjectFact(funcObj, fact)
}

// getParamPosition finds the position of the given parameter in the given function.
// Returns -1 if the parameter was not found.
func getParamPosition(funcType *ast.FuncType, param *ast.Ident) int {
	if param == nil || param.Obj == nil {
		return -1
	}

	position := 0
	for _, paramGroup := range funcType.Params.List {
		if len(paramGroup.Names) == 0 {
			position++
			continue
		}

		for _, paramDefinition := range paramGroup.Names {
			if paramDefinition.Obj == param.Obj {
				return position
			}
			position++
		}
	}

	return -1
}

// getParamIdent finds the ident of the given parameter position in the given function.
// Returns nil if the parameter was not found.
func getParamIdent(funcType *ast.FuncType, paramPosition int) *ast.Ident {
	if paramPosition < 0 {
		return nil
	}

	position := 0
	for _, paramGroup := range funcType.Params.List {
		if len(paramGroup.Names) == 0 {
			if position == paramPosition {
				return nil
			}
			position++
			continue
		}

		for _, paramDefinition := range paramGroup.Names {
			if position == paramPosition {
				return paramDefinition
			}
			position++
		}
	}

	return nil
}
