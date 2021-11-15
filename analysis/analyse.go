package analysis

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"sort"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/astutil"
	"golang.org/x/tools/go/types/typeutil"

	"github.com/serum-errors/go-serum-analyzer/analysis/scc"
)

var logf = fmt.Printf

// var logf = func(_ string, _ ...interface{}) {}

var cliArguments = struct{ requireErrorCodes bool }{}

func init() {
	Analyzer.Flags.BoolVar(&cliArguments.requireErrorCodes, "strict", false, "if this flag is set, exported error returning functions are required to declare error codes")
}

var Analyzer = &analysis.Analyzer{
	Name:     "serum",
	Doc:      "Checks that any function that has a structured docstring enumerating Serum-style error codes is telling the truth.",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      runVerify,
	FactTypes: []analysis.Fact{
		new(ErrorCodes),
		new(ErrorConstructor),
		new(ErrorType),
		new(ErrorInterface),
	},
}

type (
	ErrorCodes struct {
		Codes CodeSet
	}

	// ErrorConstructor is a fact that is used to tag functions that are error constructors,
	// meaning they take an error code parameter (string) and return an error.
	//
	// For example a constructor function "NewError(code, message string) error { return &Error{code, message} }"
	// gets an ErrorConstructor{CodeParamPosition: 0} fact.
	ErrorConstructor struct {
		CodeParamPosition int
	}
)

func (*ErrorCodes) AFact() {}

func (e *ErrorCodes) String() string {
	codes := e.Codes.Slice()
	sort.Strings(codes)
	return fmt.Sprintf("ErrorCodes: %v", strings.Join(codes, " "))
}

func (*ErrorConstructor) AFact() {}

func (e *ErrorConstructor) String() string {
	return fmt.Sprintf("ErrorConstructor: {CodeParamPosition:%d}", e.CodeParamPosition)
}

type (
	context struct {
		pass     *analysis.Pass
		lookup   *funcLookup
		scc      scc.State
		comments ast.CommentMap
	}

	funcCodesMap map[*ast.FuncDecl]funcCodes

	funcCodes struct {
		codes CodeSet
		param *funcCodeParam
	}

	funcCodeParam struct {
		ident    *ast.Ident
		position int
	}

	// funcDefinition is used to hold either an ast.FuncDecl or ast.FuncLit but not both at the same time.
	funcDefinition struct {
		funcDecl *ast.FuncDecl
		funcLit  *ast.FuncLit
	}

	funcDeclOrLit interface {
		ast.Node
	}
)

func (f *funcDefinition) node() funcDeclOrLit {
	if f.funcDecl != nil {
		return f.funcDecl
	}
	return f.funcLit
}

func (f *funcDefinition) body() *ast.BlockStmt {
	if f.funcDecl != nil {
		return f.funcDecl.Body
	}
	return f.funcLit.Body
}

func (f *funcDefinition) Type() *ast.FuncType {
	if f.funcDecl != nil {
		return f.funcDecl.Type
	}
	return f.funcLit.Type
}

func runVerify(pass *analysis.Pass) (interface{}, error) {
	lookup := collectFunctions(pass)
	comments := createCommentMap(pass)

	findAndTagErrorTypes(pass, lookup)

	interfaces := findErrorReturningInterfaces(pass)
	exportInterfaceFacts(pass, interfaces)

	funcsToAnalyse := findErrorReturningFunctions(pass, lookup)

	// Out of funcsToAnalyse get all functions that declare error codes and the actual codes they declare.
	// In the remaining analysis we only look at the functions that declare error codes or get called by an analysed function.
	funcClaims := findClaimedErrorCodes(pass, funcsToAnalyse)
	exportErrorConstructorFacts(pass, funcClaims)

	// Okay -- let's look at the functions that have made claims about their error codes.
	// We'll explore deeply to find everything that can actually affect their error return value.
	// When we reach data initialization... we look at if those types implement coded errors, and try to figure out what their actual code value is.
	// When we reach other function calls that declare their errors, that's good enough info (assuming they're also being checked for truthfulness).
	// Anything else is trouble.
	scc := scc.StartSCC() // SCC for handling of recursive functions
	c := &context{pass, lookup, scc, comments}
	for funcDecl, claims := range funcClaims {
		foundCodes, ok := lookup.foundCodes[funcDecl]
		if !ok {
			foundCodes = findErrorCodesInFunc(c, &funcDefinition{funcDecl, nil})
		}

		reportIfCodesDoNotMatch(pass, funcDecl, foundCodes, claims.codes)
	}

	// Export all claimed error codes as facts.
	// Missing error code docs or unused ones will get reported in the respective functions,
	// but on caller site only the documented behaviour matters.
	exportErrorCodeFacts(pass, funcClaims)

	findConversionsToErrorReturningInterfaces(c)

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
	types.NewFunc(token.NoPos, nil, "Cause", types.NewSignature(nil, nil, types.NewTuple(types.NewVar(token.NoPos, nil, "", types.Universe.Lookup("error").Type())), false)),
}, nil).Complete()

// isErrorCodeValid checks if the given error code is valid.
//
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

func checkErrorCodeValid(code string) error {
	if !isErrorCodeValid(code) {
		return fmt.Errorf("should match [a-zA-Z][a-zA-Z0-9\\-]*[a-zA-Z0-9]")
	}
	return nil
}

// isMethod checks if funcDecl is a method by looking if it has a single receiver.
func isMethod(funcDecl *ast.FuncDecl) bool {
	return funcDecl != nil && funcDecl.Recv != nil && len(funcDecl.Recv.List) == 1
}

// getNamedType casts the given type to *types.Named if possible,
// unpacking pointers if they occur.
// getNamedType returns nil, if said conversion fails.
func getNamedType(typ types.Type) *types.Named {
	named, ok := typ.(*types.Named)
	if ok {
		return named
	}

	pointer, ok := typ.(*types.Pointer)
	if ok {
		return getNamedType(pointer.Elem())
	}

	return nil
}

// getUnderlyingType gets the underlying type, if the passed type was a named type.
func getUnderlyingType(typ types.Type) types.Type {
	named := getNamedType(typ)
	if named == nil {
		return typ
	} else {
		return named.Underlying()
	}
}

// createCommentMap creates a map for fast access to all comments of any node in the package.
func createCommentMap(pass *analysis.Pass) ast.CommentMap {
	comments := make(ast.CommentMap)

	for _, file := range pass.Files {
		fileComments := ast.NewCommentMap(pass.Fset, file, file.Comments)
		for node, comment := range fileComments {
			comments[node] = comment
		}
	}

	return comments
}

// findErrorDocs looks at the given comments and tries to find error code declarations.
func findErrorDocs(comments *ast.CommentGroup) (CodeSet, string, bool, error) {
	if comments == nil {
		return nil, "", false, nil
	}
	return findErrorDocsSM{}.run(comments.Text())
}

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
		if checkFunctionReturnsError(pass, funcDecl.Type) {
			funcsToAnalyse = append(funcsToAnalyse, funcDecl)
		}
	})

	return funcsToAnalyse
}

// checkFunctionReturnsError determines if the given type is a function that returns an error.
// If the last result is not an error but one of the other results is, it emits a diagnostic.
func checkFunctionReturnsError(pass *analysis.Pass, funcType *ast.FuncType) bool {
	resultsList := funcType.Results
	if resultsList == nil {
		return false
	}

	lastResult := resultsList.List[len(resultsList.List)-1]
	typ := pass.TypesInfo.TypeOf(lastResult.Type)
	if !types.Implements(typ, tError) {
		// Emit diagnostic if an error is returned as non-last argument
		for _, result := range resultsList.List {
			typ := pass.TypesInfo.TypeOf(result.Type)
			if types.Implements(typ, tError) {
				pass.ReportRangef(result, "error should be returned as the last argument")
			}
		}
		return false
	}

	return true
}

// findClaimedErrorCodes finds the error codes claimed by the given functions,
// and emits diagnostics if a function does not claim error codes or
// if the format of the docstring does not match the expected format.
func findClaimedErrorCodes(pass *analysis.Pass, funcsToAnalyse []*ast.FuncDecl) funcCodesMap {
	result := funcCodesMap{}
	for _, funcDecl := range funcsToAnalyse {
		codes, errorCodeParamName, declaredNoCodesOk, err := findErrorDocs(funcDecl.Doc)
		if err != nil {
			pass.Reportf(funcDecl.Pos(), "function %q has odd docstring: %s", funcDecl.Name.Name, err)
			continue
		}

		errorCodeParam, ok := findErrorCodeParamIdent(pass, funcDecl.Type, errorCodeParamName)
		if !ok {
			continue
		}

		if len(codes) == 0 && !declaredNoCodesOk && errorCodeParam == nil {
			// Exclude Cause() methods of error types from having to declare error codes.
			// If a Cause() method declares error codes, treat it like every other method.
			if isMethod(funcDecl) {
				receiverType := pass.TypesInfo.TypeOf(funcDecl.Recv.List[0].Type)
				if types.Implements(receiverType, tReeErrorWithCause) && funcDecl.Name.Name == "Cause" {
					continue
				}
			}

			// Warn directly about any functions that are exported if they return errors,
			// but don't declare error codes in their docs.
			if cliArguments.requireErrorCodes && funcDecl.Name.IsExported() {
				pass.Reportf(funcDecl.Pos(), "function %q is exported, but does not declare any error codes", funcDecl.Name.Name)
			}
		} else {
			result[funcDecl] = funcCodes{codes, errorCodeParam}
		}
	}

	return result
}

// findErrorCodeParamIdent tries to find the error code param identifier in the parameter list
// of the given function using the name of the parameter.
func findErrorCodeParamIdent(pass *analysis.Pass, funcType *ast.FuncType, errorCodeParamName string) (*funcCodeParam, bool) {
	if errorCodeParamName == "" {
		return nil, true
	}

	position := 0
	for _, param := range funcType.Params.List { // Params is never nil
		for _, paramIdent := range param.Names {
			if paramIdent.Name != errorCodeParamName {
				position++
				continue
			}

			basic, ok := pass.TypesInfo.TypeOf(paramIdent).(*types.Basic)
			if !ok || basic.Name() != "string" {
				pass.ReportRangef(paramIdent, "error code parameter %q has to be of type string", errorCodeParamName)
				return nil, false
			}

			return &funcCodeParam{paramIdent, position}, true
		}
	}

	pass.Reportf(funcType.Pos(), "declared error code parameter %q could not be found in parameter list", errorCodeParamName)
	return nil, false
}

// exportErrorConstructorFacts exports all error code params for each function in the given map as facts.
func exportErrorConstructorFacts(pass *analysis.Pass, codes funcCodesMap) {
	for funcDecl, funcCodes := range codes {
		if funcCodes.param != nil {
			exportErrorConstructorFact(pass, funcDecl.Name, funcCodes.param)
		}
	}
}

// exportErrorConstructorFact exports the error code param for the given function as an ErrorConstructor fact.
func exportErrorConstructorFact(pass *analysis.Pass, funcIdent *ast.Ident, param *funcCodeParam) {
	definition, ok := pass.TypesInfo.Defs[funcIdent]
	if !ok {
		logf("Could not find definition for function %q!", funcIdent.Name)
		return
	}

	fn, ok := definition.(*types.Func)
	if !ok {
		logf("Definition for given identifier %q is not a function!", funcIdent.Name)
		return
	}

	fact := &ErrorConstructor{param.position}
	pass.ExportObjectFact(fn, fact)
}

// exportErrorCodeFacts exports all codes for each function in the given map as facts.
func exportErrorCodeFacts(pass *analysis.Pass, codes funcCodesMap) {
	for funcDecl, funcCodes := range codes {
		exportErrorCodesFact(pass, funcDecl.Name, funcCodes.codes)
	}
}

// exportErrorCodesFact exports all given codes for the given function as an ErrorCodes fact.
func exportErrorCodesFact(pass *analysis.Pass, funcIdent *ast.Ident, codes CodeSet) {
	definition, ok := pass.TypesInfo.Defs[funcIdent]
	if !ok {
		logf("Could not find definition for function %q!", funcIdent.Name)
		return
	}

	fn, ok := definition.(*types.Func)
	if !ok {
		logf("Definition for given identifier %q is not a function!", funcIdent.Name)
		return
	}

	fact := &ErrorCodes{codes}
	pass.ExportObjectFact(fn, fact)
}

// checkIfErrorCodesMatch checks if the two given code sets match and
// generates error messages if they don't match.
func checkIfErrorCodesMatch(foundCodes CodeSet, claimedCodes CodeSet) (bool, string) {
	missingCodes := Difference(foundCodes, claimedCodes).Slice()
	unusedCodes := Difference(claimedCodes, foundCodes).Slice()
	var errorMessages []string

	if len(missingCodes) > 0 {
		sort.Strings(missingCodes)
		errorMessages = append(errorMessages, fmt.Sprintf("missing codes: %v", missingCodes))
	}

	if len(unusedCodes) > 0 {
		sort.Strings(unusedCodes)
		errorMessages = append(errorMessages, fmt.Sprintf("unused codes: %v", unusedCodes))
	}

	errorCodesMatch := len(errorMessages) == 0
	errorMessage := strings.Join(errorMessages, " ")
	return errorCodesMatch, errorMessage
}

// reportIfCodesDoNotMatch emits a diagnostic if the given code collections don't match.
func reportIfCodesDoNotMatch(pass *analysis.Pass, funcDecl *ast.FuncDecl, foundCodes CodeSet, claimedCodes CodeSet) {
	errorCodesMatch, errorMessage := checkIfErrorCodesMatch(foundCodes, claimedCodes)
	if !errorCodesMatch {
		pass.Reportf(funcDecl.Pos(), "function %q has a mismatch of declared and actual error codes: %s", funcDecl.Name.Name, errorMessage)
	}
}

// findErrorCodesInFunc finds error codes that are returned by the given function.
// The result is also stored in the foundCodes cache of the given funcLookup.
func findErrorCodesInFunc(c *context, function *funcDefinition) CodeSet {
	pass, scc, lookup := c.pass, c.scc, c.lookup

	scc.Visit(function.node())
	result := Set()
	visitedIdents := map[*ast.Object]struct{}{}

	paramCodes := ectractErrorCodesFromConstructor(c, function)
	result = Union(result, paramCodes)

	returnCodes := findErrorCodesInFunctionReturnStmts(c, visitedIdents, function)
	result = Union(result, returnCodes)

	assignedCodes := findCodesAssignedToErrorCodeFields(pass, function, visitedIdents)
	result = Union(result, assignedCodes)

	lookup.foundCodes[function.node()] = result

	isComponentRoot, component := scc.EndVisit(function.node())
	if isComponentRoot {
		return unifyAnalysisResultForComponent(lookup, component)
	}

	return result
}

// findErrorCodesInFunctionReturnStmts looks at all return statement of the given (error returning) function
// and figures out which error codes may be returned by that statement.
func findErrorCodesInFunctionReturnStmts(c *context, visitedIdents map[*ast.Object]struct{}, function *funcDefinition) CodeSet {
	result := Set()

	ast.Inspect(function.body(), func(node ast.Node) bool {
		switch stmt := node.(type) {
		case *ast.FuncLit:
			return false // We don't want to see return statements from in a nested function right now.
		case *ast.ReturnStmt:
			annotations := getReturnStmtAnnotations(c, stmt)
			if annotations != nil && annotations.shouldOverwrite {
				result = Union(result, annotations.overwrite)
				return false
			}

			returnCodes := findErrorCodesInReturnStmt(c, visitedIdents, stmt, function)
			if annotations != nil {
				returnCodes = Difference(returnCodes, annotations.subCodes)
				returnCodes = Union(returnCodes, annotations.addCodes)
			}

			result = Union(result, returnCodes)
			return false
		}
		return true
	})

	return result
}

// unifyAnalysisResultForComponent sets the analysis result of each function in the given component to a combined result,
// containing all the error codes and affectors that result from the analysis of the individual functions.
func unifyAnalysisResultForComponent(lookup *funcLookup, component scc.Component) CodeSet {
	result := Set()

	// Create unified result using all individual results of the functions in the component.
	for _, element := range component {
		funcDecl := element.(funcDeclOrLit)
		codes := lookup.foundCodes[funcDecl]

		// lookup.analysisResults[funcDecl] will be overwritten in the next step, so using combineInplace is fine.
		result = Union(result, codes)
	}

	// Set the unified result to all functions in the component.
	for _, element := range component {
		funcDecl := element.(funcDeclOrLit)
		lookup.foundCodes[funcDecl] = result
	}

	return result
}

// findErrorCodesInReturnStmt finds all error codes that originate from the given return statement.
//
// If the return statement results list is empty (i.e. `return`), then the error codes are gathered from
// the taint spread of the named return variable for the error.
func findErrorCodesInReturnStmt(c *context, visitedIdents map[*ast.Object]struct{}, stmt *ast.ReturnStmt, function *funcDefinition) CodeSet {
	// stmt.Results can also be nil, in which case you have to look back at vars in the func sig.
	var resultExpression ast.Expr
	if len(stmt.Results) == 0 {
		resultTypes := function.Type().Results.List
		if len(resultTypes) == 0 {
			panic("should be unreachable: we already know that the function signature contains an error result.")
		}

		resultIdents := resultTypes[len(resultTypes)-1].Names
		if len(resultIdents) == 0 {
			panic("should be unreachable: an empty return statement requires either empty result list or named results.")
		}

		resultExpression = resultIdents[len(resultIdents)-1]
	} else {
		resultExpression = stmt.Results[len(stmt.Results)-1]
	}

	if resultExpression != nil {
		return findErrorCodesInExpression(c, visitedIdents, resultExpression, function)
	}
	return nil
}

// findErrorCodesInExpression finds all error codes that originate from the given expression.
func findErrorCodesInExpression(c *context, visitedIdents map[*ast.Object]struct{}, expr ast.Expr, startingFunc *funcDefinition) CodeSet {
	pass, lookup := c.pass, c.lookup

	// This can go a lot of ways:
	// - You can have a plain `*ast.Ident` (aka returning a variable).
	// - You can have an `*ast.SelectorExpr` (returning a variable from in a structure).
	// - You can have an `*ast.CallExpr` (aka returning the result of a function call).
	// - You can have an `*ast.UnaryExpr` (probably about to be an '&' and then a structure literal, but could be other things too...).
	// - This is probably not an exhaustive list...
	switch expr := astutil.Unparen(expr).(type) {
	case *ast.CallExpr:
		return findErrorCodesInCallExpression(c, expr, startingFunc)
	case *ast.Ident:
		return findErrorCodesFromIdentTaint(c, visitedIdents, expr, startingFunc)
	case *ast.UnaryExpr:
		// This might be creating a pointer, which might fulfill the error interface.  If so, we're done (and it's important to remember the pointerness).
		if expr.Op == token.AND && types.Implements(pass.TypesInfo.TypeOf(expr), tError) {
			if ident, ok := astutil.Unparen(expr.X).(*ast.Ident); ok {
				return findErrorCodesFromIdentTaint(c, visitedIdents, ident, startingFunc)
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
		pass.ReportRangef(expr, "expression is not supported in error code analysis")
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
//   - a CallExpr that targets a function literal
func findErrorCodesInCallExpression(c *context, callExpr *ast.CallExpr, startingFunc *funcDefinition) CodeSet {
	callee := typeutil.Callee(c.pass.TypesInfo, callExpr)
	return findErrorCodesFromFunctionCall(c, startingFunc, callExpr.Fun, callee, callExpr)
}

// findErrorCodesFromFunctionCall finds error codes that originate from the given function or method if it was called.
//
// The provided callExpr can be nil if no respective *ast.CallExpr exists.
func findErrorCodesFromFunctionCall(c *context, startingFunc *funcDefinition, calledFunction ast.Expr, callee types.Object, callExpr *ast.CallExpr) CodeSet {
	pass, lookup, scc := c.pass, c.lookup, c.scc

	// Get codes that originate from the callExpr itself: e.g. test-error when calling NewError("test-error")
	result := Set()
	code, ok := extractErrorCodeFromConstructorCall(pass, startingFunc, calledFunction, callee, callExpr)
	if ok {
		result.Add(code)
	}

	// We first look if the error codes are already computed and stored as a fact.
	// If so we use those, otherwise we try to recurse and compute error codes for that function.
	var fact ErrorCodes
	if callee != nil && pass.ImportObjectFact(callee, &fact) {
		return Union(result, fact.Codes)
	}

	// Get codes that originate from call expressions that are actually conversions to an error type.
	if callExpr != nil && callee == nil {
		conversionCodes := extractErrorCodesFromTypeConversion(pass, callExpr)
		if len(conversionCodes) > 0 {
			return Union(result, conversionCodes)
		}
	}

	calledFuncDef := funcDefinition{nil, nil}

	switch calledExpression := astutil.Unparen(calledFunction).(type) {
	case *ast.Ident: // this is what calls in your own package look like.
		if calledExpression.Obj == nil {
			function, ok := lookup.functions[calledExpression.Name]

			if ok {
				calledFuncDef.funcDecl = function
			} else {
				pass.ReportRangef(calledExpression, "function %q in dot-imported package does not declare error codes", calledExpression.Name)
				return Set()
			}
		} else {
			switch funcDecl := calledExpression.Obj.Decl.(type) {
			case *ast.FuncDecl: // Noramal function call
				calledFuncDef.funcDecl = funcDecl
			case *ast.TypeSpec: // Type conversion
				if callExpr != nil {
					return extractErrorCodesFromAffector(pass, lookup, startingFunc, callExpr)
				} else {
					return Set()
				}
			default: // Lambda function call (e.g. *ast.ValueSpec, *ast.AssignStmt)
				return findErrorCodesFromAllAssignedLambdas(c, calledExpression, startingFunc)
			}
		}
	case *ast.SelectorExpr: // this is what calls to other packages look like. (but can also be method call on a type)
		if target, ok := astutil.Unparen(calledExpression.X).(*ast.Ident); ok {
			if obj, ok := pass.TypesInfo.ObjectOf(target).(*types.PkgName); ok {
				// We're calling a function in a package that does not have declared error codes
				pass.ReportRangef(calledExpression, "function %q in package %q does not declare error codes", calledExpression.Sel.Name, obj.Imported().Name())
				return Set()
			}
		}

		// This case is gonna be harder than functions: We need to figure out which function declaration applies,
		// because there is no object information provided for methods calls.
		selection := pass.TypesInfo.Selections[calledExpression]
		calledFuncDef.funcDecl = lookup.searchMethod(pass, selection.Recv(), calledExpression.Sel.Name)
	case *ast.FuncLit:
		calledFuncDef.funcLit = calledExpression
	default:
		pass.ReportRangef(calledExpression, "invalid error source: definition of the unnamed function could not be found")
		return Set()
	}

	if calledFuncDef.funcDecl != nil || calledFuncDef.funcLit != nil {
		shouldRecurse := scc.HandleEdge(startingFunc.node(), calledFuncDef.node())
		if shouldRecurse {
			newCodes := findErrorCodesInFunc(c, &calledFuncDef)
			result = Union(result, newCodes)
			scc.AfterRecurse(startingFunc.node(), calledFuncDef.node())
		} else if cachedResult, ok := lookup.foundCodes[calledFuncDef.node()]; ok {
			result = Union(result, cachedResult)
		}
	} else {
		// Could e.g. be a method which is defined in another package
		pass.ReportRangef(calledFunction, "called function does not declare error codes")
	}

	return result
}

// findErrorCodesFromAllAssignedLambdas finds error codes in the given function,
// by looking into the definition of all lambdas directly or indirectly assigned to the given identifier.
func findErrorCodesFromAllAssignedLambdas(c *context, ident *ast.Ident, function *funcDefinition) CodeSet {
	pass := c.pass

	taintResult := taintSpreadForIdentOfImmutableType(pass, map[*ast.Object]struct{}{}, ident, function)

	for _, badIdent := range taintResult.identOutOfScope {
		if function.funcDecl != nil { // expression is inside a function
			pass.ReportRangef(badIdent, "error returning function literal may not be a parameter, receiver or global variable")
		} else { // expression is inside a lambda (function literal)
			pass.ReportRangef(badIdent, "error returning function literal may not be a parameter, global variable or other variables declared outside of the function body")
		}
	}

	for _, destruct := range taintResult.destructAssignment {
		pass.ReportRangef(destruct.source, "unsupported: assigning result of function call to variable %q is not allowed", destruct.target.Name)
	}

	result := Set()
	for _, expr := range taintResult.expressions {
		newCodes := findErrorCodesInLambdaAssignment(c, ident, expr, function)
		result = Union(result, newCodes)
	}
	return result
}

func findErrorCodesInLambdaAssignment(c *context, ident *ast.Ident, assignedExpr ast.Expr, function *funcDefinition) CodeSet {
	pass := c.pass
	var result CodeSet

	switch rhsEntry := astutil.Unparen(assignedExpr).(type) {
	case *ast.FuncLit:
		result = findErrorCodesInFunc(c, &funcDefinition{nil, rhsEntry})
	case *ast.Ident: // name of a function
		callee := pass.TypesInfo.Uses[rhsEntry]
		result = findErrorCodesFromFunctionCall(c, function, rhsEntry, callee, nil)
	case *ast.SelectorExpr: // name of a function in other package
		var callee types.Object
		if sel, ok := pass.TypesInfo.Selections[rhsEntry]; ok {
			callee = sel.Obj()
		} else {
			callee = pass.TypesInfo.Uses[rhsEntry.Sel]
		}
		result = findErrorCodesFromFunctionCall(c, function, rhsEntry, callee, nil)
	default:
		pass.ReportRangef(rhsEntry, "unsupported: assignment to variable %q can only be an identifier or function literal", ident.Name)
	}

	return result
}

// findErrorCodesFromIdentTaint finds error codes in the given function, by tracking all assignments to the given ident within the function.
func findErrorCodesFromIdentTaint(c *context, visitedIdents map[*ast.Object]struct{}, ident *ast.Ident, function *funcDefinition) CodeSet {
	pass := c.pass

	taintResult := taintSpreadForIdentAllowLeak(pass, visitedIdents, ident, function)

	for _, badIdent := range taintResult.identOutOfScope {
		if function.funcDecl != nil { // expression is inside a function
			pass.ReportRangef(badIdent, "returned error may not be a parameter, receiver or global variable")
		} else { // expression is inside a lambda (function literal)
			pass.ReportRangef(badIdent, "returned error may not be a parameter, global variable or other variables declared outside of the function body")
		}
	}

	result := Set()

	for _, expr := range taintResult.expressions {
		newCodes := findErrorCodesInExpression(c, visitedIdents, expr, function)
		result = Union(result, newCodes)
	}

	for _, destruct := range taintResult.destructAssignment {
		callExpr, ok := astutil.Unparen(destruct.source).(*ast.CallExpr)
		if !ok {
			panic("should be unreachable: destructirung assignment should only originate from a call expression.")
		}

		funType, ok := pass.TypesInfo.TypeOf(callExpr.Fun).(*types.Signature)
		if !ok {
			panic("should be unreachable: function of call expression should always be of type signature.")
		}

		// Destructuring mode.
		// We're going to make some crass simplifications here, and say... if this is anything other than the last arg, you're not supported.
		if destruct.position != funType.Results().Len()-1 {
			pass.ReportRangef(destruct.target, "unsupported: tracking error codes for function call with error as non-last return argument")
			continue
		}

		newCodes := findErrorCodesInCallExpression(c, callExpr, function)
		result = Union(result, newCodes)
	}

	return result
}

// isIdentOriginOutsideFunctionScope checks if the origin of the given ident is outside of the scope of the given function.
func isIdentOriginOutsideFunctionScope(function *funcDefinition, ident *ast.Ident) bool {
	if ident.Name == "nil" {
		return false
	}

	functionPos := function.body().Pos()
	if function.Type().Results != nil {
		// Results are allowed too, because named results may be declared there.
		functionPos = function.Type().Results.Pos()
	}

	return ident.Obj == nil ||
		ident.Obj.Pos() <= functionPos ||
		ident.Obj.Pos() >= function.body().End()
}

func findCodesAssignedToErrorCodeFields(pass *analysis.Pass, function *funcDefinition, errorIdents map[*ast.Object]struct{}) CodeSet {
	result := Set()

	for errorIdent := range errorIdents {
		newCodes := findCodesAssignedToErrorCodeField(pass, function, nil, errorIdent)
		result = Union(result, newCodes)
	}

	return result
}

func findCodesAssignedToErrorCodeField(pass *analysis.Pass, function *funcDefinition, errorType *ErrorType, errorIdent *ast.Object) CodeSet {
	result := Set()

	if errorIdent == nil {
		return result
	}

	ast.Inspect(function.node(), func(node ast.Node) bool {
		assignment, ok := node.(*ast.AssignStmt)
		if !ok {
			return true
		}

		newCodes := findCodesAssignedToErrorCodeFieldInAssignment(pass, function, errorType, errorIdent, assignment)
		result = Union(result, newCodes)

		return false
	})

	return result
}

// findCodesAssignedToErrorCodeField searches through the given assignment and returns every constant code assigned to the error code field.
// For invalid assignments to the error code field, diagnostics are emitted.
func findCodesAssignedToErrorCodeFieldInAssignment(pass *analysis.Pass, function *funcDefinition, errorType *ErrorType, errorIdent *ast.Object, assignment *ast.AssignStmt) CodeSet {
	result := Set()

	if errorIdent == nil {
		return result
	}

	for i, lhsEntry := range assignment.Lhs {
		lhsEntry, ok := lhsEntry.(*ast.SelectorExpr)
		if !ok {
			continue
		}

		objIdent, ok := lhsEntry.X.(*ast.Ident)
		if !ok || objIdent.Obj == nil {
			continue // Cannot inspect assignments to more complicated expressions. (yet?)
		}

		if objIdent.Obj != errorIdent {
			continue // Not the ident we're looking for.
		}

		// Found an assignment to a field of the error we're looking at.
		// Try to get the error type for the ident to see if the assignment is to the error code field.
		if errorType == nil {
			var err error
			errorType, err = getErrorTypeForError(pass, pass.TypesInfo.Types[objIdent].Type)
			if err != nil || errorType == nil || errorType.Field == nil {
				continue
			}
		}

		// Found valid error type, that has a error code field defined:
		// Check if fields match and if they do try to get the error code from the assignment.
		if errorType.Field.Name != lhsEntry.Sel.Name {
			continue
		}

		if len(assignment.Lhs) != len(assignment.Rhs) {
			pass.ReportRangef(assignment.Rhs[0], "error code has to be constant value or error code parameter")
			continue
		}

		code, ok := extractErrorCodeFromStringExpression(pass, function, assignment.Rhs[i])
		if ok {
			result.Add(code)
		}
	}

	return result
}

// checkErrorTypeHasLegibleCode makes sure that the `Code() string` function
// on a type either returns a constant or a single struct field.
// If you want to write your own error types and have them be recognized, it should be this simple.
func checkErrorTypeHasLegibleCode(pass *analysis.Pass, seen ast.Expr) bool { // probably should return a lookup function.
	typ := pass.TypesInfo.TypeOf(seen)
	return types.Implements(typ, tReeError) || types.Implements(types.NewPointer(typ), tReeError)
}
