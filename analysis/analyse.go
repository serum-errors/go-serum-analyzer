package analysis

import (
	"fmt"
	"go/ast"
	"go/constant"
	"go/token"
	"go/types"
	"sort"
	"strconv"
	"strings"

	"github.com/warpfork/go-ree/analysis/scc"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/types/typeutil"
)

var logf = fmt.Printf

// var logf = func(_ string, _ ...interface{}) {}

var VerifyAnalyzer = &analysis.Analyzer{
	Name:     "reeverify",
	Doc:      "Checks that any function that has a ree-style docstring enumerating error codes is telling the truth.",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      runVerify,
	FactTypes: []analysis.Fact{
		new(ErrorCodes),
		new(ErrorType),
	},
}

type ErrorCodes struct {
	Codes []string
}

func (*ErrorCodes) AFact() {}

func (e *ErrorCodes) String() string {
	sort.Strings(e.Codes)
	return fmt.Sprintf("ErrorCodes: %v", strings.Join(e.Codes, " "))
}

// ErrorType is a fact about a ree.Error type,
// declaring which error codes Code() might return,
// and/or what field gets returned by a call to Code().
type ErrorType struct {
	Codes []string        // error codes, or nil
	Field *ErrorCodeField // field information, or nil
}

// ErrorCodeField is part of ErrorType,
// and declares the field that might be returned by the Code() method of the ree.Error.
type ErrorCodeField struct {
	Name     string
	Position int
}

func (*ErrorType) AFact() {}

func (e *ErrorType) String() string {
	sort.Strings(e.Codes)
	return fmt.Sprintf("ErrorType{Field:%v, Codes:%v}", e.Field, strings.Join(e.Codes, " "))
}

func (f *ErrorCodeField) String() string {
	return fmt.Sprintf("{Name:%q, Position:%d}", f.Name, f.Position)
}

type funcCodes map[*ast.FuncDecl]codeSet

// isErrorCodeValid checks if the given error code is valid.
// Valid error codes have to match against: "^[a-zA-Z][a-zA-Z0-9\-]*[a-zA-Z0-9]$" or "^[a-zA-Z]$".
func isErrorCodeValid(code string) bool {
	if len(code) == 0 {
		return false
	}

	// Verify that first and last char do not contain invalid values.
	if code[0] == '-' || (code[0] >= '0' && code[0] <= '9') {
		return false
	}
	if code[len(code)-1] == '-' {
		return false
	}

	// Verify that the remaining chars match [a-zA-Z0-9\-]
	for _, c := range code {
		if !(c == '-' || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9')) {
			return false
		}
	}

	return true
}

// isMethod checks if funcDecl is a method by looking if it has a single receiver.
func isMethod(funcDecl *ast.FuncDecl) bool {
	return funcDecl != nil && funcDecl.Recv != nil && len(funcDecl.Recv.List) == 1
}

func runVerify(pass *analysis.Pass) (interface{}, error) {
	lookup := collectFunctions(pass)

	findAndTagErrorTypes(pass, lookup)

	funcsToAnalyse := findErrorReturningFunctions(pass, lookup)

	// Out of funcsToAnalyse get all functions that declare error codes and the actual codes they declare.
	// In the remaining analysis we only look at the functions that declare error codes or get called by an analysed function.
	funcClaims := findClaimedErrorCodes(pass, funcsToAnalyse)

	// Okay -- let's look at the functions that have made claims about their error codes.
	// We'll explore deeply to find everything that can actually affect their error return value.
	// When we reach data initialization... we look at if those types implement coded errors, and try to figure out what their actual code value is.
	// When we reach other function calls that declare their errors, that's good enough info (assuming they're also being checked for truthfulness).
	// Anything else is trouble.
	scc := scc.StartSCC() // SCC for handling of recursive functions
	for funcDecl, claimedCodes := range funcClaims {
		foundCodes, ok := lookup.foundCodes[funcDecl]
		if !ok {
			foundCodes = findErrorCodesInFunc(pass, lookup, scc, funcDecl)
		}

		reportIfCodesDoNotMatch(pass, funcDecl, foundCodes, claimedCodes)
	}

	// Export all claimed error codes as facts.
	// Missing error code docs or unused ones will get reported in the respective functions,
	// but on caller site only the documented behaviour matters.
	exportFunctionErrorCodes(pass, funcClaims)

	return nil, nil
}

var tError = types.NewInterfaceType([]*types.Func{
	types.NewFunc(token.NoPos, nil, "Error", types.NewSignature(nil, nil, types.NewTuple(types.NewVar(token.NoPos, nil, "", types.Typ[types.String])), false)),
}, nil).Complete()

var tReeError = types.NewInterfaceType([]*types.Func{
	types.NewFunc(token.NoPos, nil, "Error", types.NewSignature(nil, nil, types.NewTuple(types.NewVar(token.NoPos, nil, "", types.Typ[types.String])), false)),
	types.NewFunc(token.NoPos, nil, "Code", types.NewSignature(nil, nil, types.NewTuple(types.NewVar(token.NoPos, nil, "", types.Typ[types.String])), false)),
}, nil).Complete()

var tReeErrorWithCause = types.NewInterfaceType([]*types.Func{
	tReeError.Method(0),
	tReeError.Method(1),
	types.NewFunc(token.NoPos, nil, "Cause", types.NewSignature(nil, nil, types.NewTuple(types.NewVar(token.NoPos, nil, "", types.NewNamed(types.NewTypeName(token.NoPos, nil, "error", tError), nil, nil))), false)),
}, nil).Complete()

// findErrorReturningFunctions looks for functions that return an error,
// and emits a diagnostic if a function returns an error, but not as the last argument.
func findErrorReturningFunctions(pass *analysis.Pass, lookup *funcLookup) []*ast.FuncDecl {
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
	lookup.forEach(func(funcDecl *ast.FuncDecl) {
		resultsList := funcDecl.Type.Results
		if resultsList == nil {
			return
		}
		lastResult := resultsList.List[len(resultsList.List)-1]
		typ := pass.TypesInfo.TypeOf(lastResult.Type)
		if !types.Implements(typ, tError) {
			// Emit diagnostic if an error is returned as non-last argument
			for _, result := range resultsList.List {
				typ := pass.TypesInfo.TypeOf(result.Type)
				if types.Implements(typ, tError) {
					pass.Reportf(result.Pos(), "error should be returned as the last argument")
				}
			}
			return
		}
		funcsToAnalyse = append(funcsToAnalyse, funcDecl)
	})

	return funcsToAnalyse
}

// findClaimedErrorCodes finds the error codes claimed by the given functions,
// and emits diagnostics if a function does not claim error codes or
// if the format of the docstring does not match the expected format.
func findClaimedErrorCodes(pass *analysis.Pass, funcsToAnalyse []*ast.FuncDecl) funcCodes {
	result := funcCodes{}
	for _, funcDecl := range funcsToAnalyse {
		codes, err := findErrorDocs(funcDecl)
		if err != nil {
			pass.Reportf(funcDecl.Pos(), "function %q has odd docstring: %s", funcDecl.Name.Name, err)
			continue
		}

		if len(codes) == 0 {
			// Exclude Cause() methods of error types from having to declare error codes.
			// If a Cause() method declares error codes, treat it like every other method.
			if isMethod(funcDecl) {
				receiverType := pass.TypesInfo.TypeOf(funcDecl.Recv.List[0].Type)
				if types.Implements(receiverType, tReeErrorWithCause) {
					continue
				}
			}

			// Warn directly about any functions that are exported if they return errors,
			// but don't declare error codes in their docs.
			if funcDecl.Name.IsExported() {
				pass.Reportf(funcDecl.Pos(), "function %q is exported, but does not declare any error codes", funcDecl.Name.Name)
			}
		} else {
			result[funcDecl] = codes
		}
	}

	return result
}

func findErrorDocs(funcDecl *ast.FuncDecl) (codeSet, error) {
	if funcDecl.Doc == nil {
		return nil, nil
	}
	return findErrorDocsSM{}.run(funcDecl.Doc.Text())
}

// exportFunctionErrorCodes exports all codes for each function in the given map as facts.
func exportFunctionErrorCodes(pass *analysis.Pass, codes funcCodes) {
	for funcDecl, codeSet := range codes {
		fn, ok := pass.TypesInfo.Defs[funcDecl.Name].(*types.Func)
		if !ok {
			logf("Could not find definition for function %q!", funcDecl.Name.Name)
			continue
		}

		fact := ErrorCodes{codeSet.slice()}
		pass.ExportObjectFact(fn, &fact)
	}
}

// extractErrorCodesFromAffector extracts all error codes from the given affectors and returns them.
func extractErrorCodesFromAffector(pass *analysis.Pass, lookup *funcLookup, funcDecl *ast.FuncDecl, affector ast.Expr) codeSet {
	result := set()

	// Make sure method "Code() string" is present
	if !checkErrorTypeHasLegibleCode(pass, affector) {
		pass.ReportRangef(affector, "expression does not define an error code")
		return result
	}

	errorType, err := getErrorTypeForError(pass, lookup, pass.TypesInfo.Types[affector].Type)
	if err != nil || errorType == nil {
		pass.ReportRangef(affector, "expression is not a valid error: error types must return constant error codes or a single field")
	}
	if err != nil {
		logf("Error while looking at affector: %v (Affector: %#v)\n", err, affector)
	} else if errorType != nil {
		if len(errorType.Codes) > 0 {
			result = union(result, sliceToSet(errorType.Codes))
		}

		if errorType.Field != nil {
			code, err := extractFieldErrorCodes(pass, affector, funcDecl, errorType)
			if err == nil {
				result.add(code)
			} else {
				pass.ReportRangef(affector, "%v", err)
			}
		}
	}

	return result
}

// reportIfCodesDoNotMatch emits a diagnostic if the given code collections don't match.
func reportIfCodesDoNotMatch(pass *analysis.Pass, funcDecl *ast.FuncDecl, foundCodes codeSet, claimedCodes codeSet) {
	missingCodes := difference(foundCodes, claimedCodes).slice()
	unusedCodes := difference(claimedCodes, foundCodes).slice()
	var errorMessages []string
	if len(missingCodes) != 0 {
		sort.Strings(missingCodes)
		errorMessages = append(errorMessages, fmt.Sprintf("missing codes: %v", missingCodes))
	}
	if len(unusedCodes) != 0 {
		sort.Strings(unusedCodes)
		errorMessages = append(errorMessages, fmt.Sprintf("unused codes: %v", unusedCodes))
	}

	if len(errorMessages) != 0 {
		errorMessage := strings.Join(errorMessages, " ")
		pass.Reportf(funcDecl.Pos(), "function %q has a mismatch of declared and actual error codes: %s", funcDecl.Name.Name, errorMessage)
	}
}

// findErrorCodesInFunc finds error codes that are returned by the given function.
// The result is also stored in the foundCodes cache of the given funcLookup.
func findErrorCodesInFunc(pass *analysis.Pass, lookup *funcLookup, scc scc.State, funcDecl *ast.FuncDecl) codeSet {
	scc.Visit(funcDecl)
	result := set()
	visitedIdents := map[*ast.Ident]struct{}{}

	ast.Inspect(funcDecl, func(node ast.Node) bool {
		switch stmt := node.(type) {
		case *ast.FuncLit:
			return false // We don't want to see return statements from in a nested function right now.
		case *ast.ReturnStmt:
			// stmt.Results can also be nil, in which case you have to look back at vars in the func sig.
			var resultExpression ast.Expr
			if len(stmt.Results) == 0 {
				resultTypes := funcDecl.Type.Results.List
				if len(resultTypes) == 0 {
					panic("Should be unreachable: we already know that the function signature contains an error result.")
				}

				resultIdents := resultTypes[len(resultTypes)-1].Names
				if len(resultIdents) == 0 {
					panic("Should be unreachable: an empty return statement requires either empty result list or named results.")
				}

				resultExpression = resultIdents[len(resultIdents)-1]
			} else {
				resultExpression = stmt.Results[len(stmt.Results)-1]
			}

			// This can go a lot of ways:
			// - You can have a plain `*ast.Ident` (aka returning a variable).
			// - You can have an `*ast.SelectorExpr` (returning a variable from in a structure).
			// - You can have an `*ast.CallExpr` (aka returning the result of a function call).
			// - You can have an `*ast.UnaryExpr` (probably about to be an '&' and then a structure literal, but could be other things too...).
			// - This is probably not an exhaustive list...
			if resultExpression != nil {
				newCodes := findErrorCodesInExpression(pass, lookup, scc, visitedIdents, resultExpression, funcDecl)
				result = union(result, newCodes)
			}

			return false
		}
		return true
	})

	lookup.foundCodes[funcDecl] = result

	isComponentRoot, component := scc.EndVisit(funcDecl)
	if isComponentRoot {
		return unifyAnalysisResultForComponent(lookup, component)
	}

	return result
}

// unifyAnalysisResultForComponent sets the analysis result of each function in the given component to a combined result,
// containing all the error codes and affectors that result from the analysis of the individual functions.
func unifyAnalysisResultForComponent(lookup *funcLookup, component scc.Component) codeSet {
	result := set()

	// Create unified result using all individual results of the functions in the component.
	for _, element := range component {
		funcDecl := element.(*ast.FuncDecl)
		codes := lookup.foundCodes[funcDecl]

		// lookup.analysisResults[funcDecl] will be overwritten in the next step, so using combineInplace is fine.
		result = union(result, codes)
	}

	// Set the unified result to all functions in the component.
	for _, element := range component {
		funcDecl := element.(*ast.FuncDecl)
		lookup.foundCodes[funcDecl] = result
	}

	return result
}

// findErrorCodesInExpression finds all error codes that originate from the given expression.
func findErrorCodesInExpression(pass *analysis.Pass, lookup *funcLookup, scc scc.State, visitedIdents map[*ast.Ident]struct{}, expr ast.Expr, startingFunc *ast.FuncDecl) codeSet {
	switch expr := expr.(type) {
	case *ast.CallExpr:
		return findErrorCodesInCallExpression(pass, lookup, scc, expr, startingFunc)
	case *ast.Ident:
		return findErrorCodesFromIdentTaint(pass, lookup, scc, visitedIdents, expr, startingFunc)
	case *ast.UnaryExpr:
		// This might be creating a pointer, which might fulfill the error interface.  If so, we're done (and it's important to remember the pointerness).
		if expr.Op == token.AND && types.Implements(pass.TypesInfo.TypeOf(expr), tError) { // TODO the docs of this function are not truthfully admitting how specific this is.
			if ident, ok := expr.X.(*ast.Ident); ok {
				return findErrorCodesFromIdentTaint(pass, lookup, scc, visitedIdents, ident, startingFunc)
			} else {
				return extractErrorCodesFromAffector(pass, lookup, startingFunc, expr)
			}
		}

		// If it's not fulfilling the error interface it's not supported
		pass.ReportRangef(expr, "expression does not implement valid error type")
		return nil
	case *ast.CompositeLit, *ast.BasicLit: // Actual value creation!
		return extractErrorCodesFromAffector(pass, lookup, startingFunc, expr)
	default:
		logf(":: findAffectorsInFunc does not yet handle %#v\n", expr)
		return nil
	}
}

// findErrorCodesInCallExpression finds error codes that originate from the given function or method call.
//
// The given CallExpr could be:
//   - a CallExpr that targets another function that has declared error codes (yay!)
//   - a CallExpr that crosses package boundaries (get declared error codes or fail)
//   - a CallExpr that's an interface (we can't really look deeper than that)
//   - a CallExpr that targets another function in this package (recurse or load from cache)
//   - a CallExpr that targets a function literal (not handled)
func findErrorCodesInCallExpression(pass *analysis.Pass, lookup *funcLookup, scc scc.State, callExpr *ast.CallExpr, startingFunc *ast.FuncDecl) codeSet {
	// For a CallExpr we first look if the error codes are already computed and stored as a fact.
	// If so we use those, otherwise we try to recurse and compute error codes for that function.
	callee := typeutil.Callee(pass.TypesInfo, callExpr)
	var fact ErrorCodes
	if callee != nil && pass.ImportObjectFact(callee, &fact) {
		return sliceToSet(fact.Codes)
	}

	var calledFunc *ast.FuncDecl

	switch calledExpression := callExpr.Fun.(type) {
	case *ast.Ident: // this is what calls in your own package look like. // TODO and dot-imported, I guess.  Yeesh.
		switch funcDecl := calledExpression.Obj.Decl.(type) {
		case *ast.FuncDecl: // Noramal function call
			calledFunc = funcDecl
		case *ast.TypeSpec: // Type conversion
			return extractErrorCodesFromAffector(pass, lookup, startingFunc, callExpr)
		}
	case *ast.SelectorExpr: // this is what calls to other packages look like. (but can also be method call on a type)
		if target, ok := calledExpression.X.(*ast.Ident); ok {
			if obj, ok := pass.TypesInfo.ObjectOf(target).(*types.PkgName); ok {
				// We're calling a function in a package that does not have declared error codes
				pass.ReportRangef(calledExpression, "function %q in package %q does not declare error codes", calledExpression.Sel.Name, obj.Imported().Name())
				return set()
			}
		}

		// This case is gonna be harder than functions: We need to figure out which function declaration applies,
		// because there is no object information provided for methods calls.
		selection := pass.TypesInfo.Selections[calledExpression]
		calledFunc = lookup.searchMethod(pass, selection.Recv(), calledExpression.Sel.Name)
	default:
		pass.ReportRangef(calledExpression, "unnamed functions are not yet supported in error code analysis")
		return set()
	}

	result := set()

	if calledFunc != nil {
		shouldRecurse := scc.HandleEdge(startingFunc, calledFunc)
		if shouldRecurse {
			result = findErrorCodesInFunc(pass, lookup, scc, calledFunc)
			scc.AfterRecurse(startingFunc, calledFunc)
		} else if cachedResult, ok := lookup.foundCodes[calledFunc]; ok {
			result = cachedResult
		}
	} else {
		// Could e.g. be a method which is defined in another package
		pass.ReportRangef(callExpr.Fun, "called function does not declare error codes")
	}

	return result
}

// findErrorCodesFromIdentTaint finds error codes in the given function, by tracking all assignments to the given ident within the function.
func findErrorCodesFromIdentTaint(pass *analysis.Pass, lookup *funcLookup, scc scc.State, visitedIdents map[*ast.Ident]struct{}, ident *ast.Ident, within *ast.FuncDecl) codeSet {
	// Mark ident as visited to avoid revisiting it again (possibly resulting in an endles loop)
	if _, ok := visitedIdents[ident]; ok {
		return nil
	}
	visitedIdents[ident] = struct{}{}

	// Check that the identifier is a local variable.
	if ident.Obj != nil {
		withinPos := within.Body.Pos()
		if within.Type.Results != nil {
			// Results are allowed too, because named results may be declared there.
			withinPos = within.Type.Results.Pos()
		}

		if ident.Obj.Pos() <= withinPos || ident.Obj.Pos() >= within.Body.End() {
			pass.ReportRangef(ident, "returned error may not be a parameter, receiver or global variable")
		}
	}

	result := set()

	// Look for for `*ast.AssignStmt` in the function that could've affected this.
	ast.Inspect(within, func(node ast.Node) bool {
		// n.b., do *not* filter out *`ast.FuncLit`: statements inside closures can assign things!
		assignment, ok := node.(*ast.AssignStmt)
		if !ok {
			return true
		}

		// Look for our ident's object in the left-hand-side of the assign.
		// Either follow up on the statement at the same index in the Rhs,
		// or watch out for a shorter Rhs that's just a CallExpr (i.e. it's a destructuring assignment).
		for i, lhsEntry := range assignment.Lhs {
			switch lhsEntry := lhsEntry.(type) {
			case *ast.Ident:
				if lhsEntry.Obj != ident.Obj {
					continue
				}

				if len(assignment.Lhs) == len(assignment.Rhs) {
					newCodes := findErrorCodesInExpression(pass, lookup, scc, visitedIdents, assignment.Rhs[i], within)
					result = union(result, newCodes)
				} else {
					// Destructuring mode.
					// We're going to make some crass simplifications here, and say... if this is anything other than the last arg, you're not supported.
					if i != len(assignment.Lhs)-1 {
						pass.ReportRangef(lhsEntry, "unsupported: tracking error codes for function call with error as non-last return argument")
						continue
					}
					// Because it's a CallExpr, we're done here: this is part of the result.
					if callExpr, ok := assignment.Rhs[0].(*ast.CallExpr); ok {
						newCodes := findErrorCodesInCallExpression(pass, lookup, scc, callExpr, within)
						result = union(result, newCodes)
					} else {
						panic("what?")
					}
				}
			case *ast.SelectorExpr:
				objIdent, ok := lhsEntry.X.(*ast.Ident)
				if !ok || objIdent.Obj == nil {
					continue // Cannot inspect assignments to more complicated expressions. (yet?)
				}

				if objIdent.Obj != ident.Obj {
					continue // Not the ident we're looking for.
				}

				// Found an assignment to a field of the error we're looking at.
				// Try to get the error type for the ident to see if the assignment is to the error code field.
				errorType, err := getErrorTypeForError(pass, lookup, pass.TypesInfo.Types[objIdent].Type)
				if err != nil || errorType == nil || errorType.Field == nil {
					continue
				}

				// Found valid error type, that has a error code field defined:
				// Check if fields match and if they do try to get the error code from the assignment.
				if errorType.Field.Name != lhsEntry.Sel.Name {
					continue
				}

				if len(assignment.Lhs) != len(assignment.Rhs) {
					pass.ReportRangef(lhsEntry, "error code field has to be assigned a constant value")
					continue
				}

				value := pass.TypesInfo.Types[assignment.Rhs[i]].Value
				if value == nil {
					pass.ReportRangef(lhsEntry, "error code field has to be assigned a constant value")
					continue
				}

				code, err := getErrorCodeFromConstant(value)
				if err != nil {
					pass.ReportRangef(assignment.Rhs[i], "%v", err)
					continue
				}

				result.add(code)
			}
		}
		return true
	})

	return result
}

// checkErrorTypeHasLegibleCode makes sure that the `Code() string` function
// on a type either returns a constant or a single struct field.
// If you want to write your own ree.Error, it should be this simple.
func checkErrorTypeHasLegibleCode(pass *analysis.Pass, seen ast.Expr) bool { // probably should return a lookup function.
	typ := pass.TypesInfo.TypeOf(seen)
	return types.Implements(typ, tReeError) || types.Implements(types.NewPointer(typ), tReeError)
}

// extractFieldErrorCodes finds a possible error code from the given constructor expression.
//
// The expression evaluates to an error of the given error type, which has its errorType.Field set to a value (not nil).
func extractFieldErrorCodes(pass *analysis.Pass, expr ast.Expr, funcDecl *ast.FuncDecl, errorType *ErrorType) (string, error) {
	if errorType == nil || errorType.Field == nil {
		return "", fmt.Errorf("cannot extract field error code without field definition")
	}

	switch expr := expr.(type) {
	case *ast.CompositeLit:
		// Key-based composite literal:
		// Use the field name to find the error code.
		for _, element := range expr.Elts {
			element, ok := element.(*ast.KeyValueExpr)
			if !ok { // Either all elements are KeyValueExpr or none.
				break
			}

			ident, ok := element.Key.(*ast.Ident)
			if !ok {
				logf("found weird key %#v in composite literal %#v\n", element.Key, expr)
				break
			}

			if errorType.Field.Name == ident.Name {
				info, ok := pass.TypesInfo.Types[element.Value]
				if ok && info.Value != nil {
					return getErrorCodeFromConstant(info.Value)
				}
			}
		}

		// Position-based composite literal:
		// Use the field position to find the error code.
		pos := errorType.Field.Position
		if pos < len(expr.Elts) {
			info, ok := pass.TypesInfo.Types[expr.Elts[pos]]
			if ok && info.Value != nil {
				return getErrorCodeFromConstant(info.Value)
			}
		}
	case *ast.UnaryExpr:
		if expr.Op == token.AND {
			return extractFieldErrorCodes(pass, expr.X, funcDecl, errorType)
		}
	default:
		logf("extractErrorCodes did not yet handle: %#v\n", expr)
	}

	return "", fmt.Errorf("error code field has to be instantiated by constant value")
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

	if !isErrorCodeValid(result) {
		return "", fmt.Errorf("error code has invalid format: should match [a-zA-Z][a-zA-Z0-9\\-]*[a-zA-Z0-9]")
	}

	return result, nil
}
