package analysis

import (
	"fmt"
	"go/ast"
	"go/types"
	"sort"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/astutil"
	"golang.org/x/tools/go/ast/inspector"
)

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

// findAndTagErrorTypes finds all errors with a Code() method
// and exports an ErrorType fact for all valid error types.
func findAndTagErrorTypes(pass *analysis.Pass, lookup *funcLookup) {
	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	// We only need to see type declarations.
	nodeFilter := []ast.Node{
		(*ast.GenDecl)(nil),
	}

	inspect.Nodes(nodeFilter, func(node ast.Node, _ bool) bool {
		genDecl := node.(*ast.GenDecl)

		for _, spec := range genDecl.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}

			typ := pass.TypesInfo.Defs[typeSpec.Name].Type()

			// Filter out all types that are not errors with a Code() method.
			if !types.Implements(typ, tReeError) {
				typ = types.NewPointer(typ)
				if !types.Implements(typ, tReeError) {
					continue
				}
			}

			// Export error type fact for error.
			err := tagErrorType(pass, lookup, typ, typeSpec)
			if err != nil {
				pass.ReportRangef(node, "%v", err)
			}
		}

		// Never recurse deeper.
		return false
	})
}

// tagErrorType exports an ErrorType fact for the given error if it's a valid error type.
func tagErrorType(pass *analysis.Pass, lookup *funcLookup, err types.Type, spec *ast.TypeSpec) error {
	namedErr := getNamedType(err)
	if namedErr == nil {
		logf("err type: %#v\n", err)
		return fmt.Errorf("type is an invalid error type")
	}

	// Ignore interface types: we don't need to tag them, only concrete implementations.
	if _, ok := namedErr.Underlying().(*types.Interface); ok {
		return nil
	}

	funcDecl, receiver := getCodeFuncFromError(pass, lookup, err)
	if funcDecl == nil {
		return fmt.Errorf(`found no method "Code() string"`)
	}
	errorType := analyseCodeMethod(pass, spec, funcDecl, receiver)

	if errorType == nil {
		return fmt.Errorf("type %q is an invalid error type: could not find any error codes", namedErr.Obj().Name())
	}

	analyseMethodsOfErrorType(pass, lookup, errorType, err)

	pass.ExportObjectFact(namedErr.Obj(), errorType)
	return nil
}

// getErrorTypeForError gets the ErrorType for the given error from cache,
// or on a cache miss computes said ErrorType and stores it in the cache.
func getErrorTypeForError(pass *analysis.Pass, lookup *funcLookup, err types.Type) (*ErrorType, error) {
	namedErr := getNamedType(err)
	if namedErr == nil {
		logf("err type: %#v\n", err)
		return nil, fmt.Errorf("passed invalid err type to getErrorTypeForError")
	}

	errorType := new(ErrorType)
	if pass.ImportObjectFact(namedErr.Obj(), errorType) {
		return errorType, nil
	}

	return nil, nil
}

// getCodeFuncFromError finds and returns the method declaration of "Code() string" for the given error type.
//
// The second result is the identifier which is the receiver of the method,
// or nil if the receiver is unnamed.
func getCodeFuncFromError(pass *analysis.Pass, lookup *funcLookup, err types.Type) (result *ast.FuncDecl, receiver *ast.Ident) {
	// Use lookup struct to find correct Code() method
	methods, ok := lookup.methods["Code"]
	if !ok {
		return nil, nil
	}

	// Search through all methods named "Code" to find the right one for the given error type.
	for _, funcDecl := range methods {
		// funcDecl is guaranteed to have one receiver, because it is a method
		receiverField := funcDecl.Recv.List[0]
		if !errorTypesSubset(pass.TypesInfo.TypeOf(receiverField.Type), err) {
			continue
		}

		if len(receiverField.Names) == 1 {
			return funcDecl, receiverField.Names[0]
		}

		return funcDecl, nil
	}

	return nil, nil
}

// errorTypesSubset checks if type1 is a subset of type2.
func errorTypesSubset(type1, type2 types.Type) bool {
	pointer2, ok2 := type2.(*types.Pointer)
	return types.Identical(type1, type2) ||
		(ok2 && types.Identical(type1, pointer2.Elem()))
}

// analyseCodeMethod inspects the error type.
//
// If the Code() method returns a constant value:
//     That is the error code we're looking for
//     Having multiple return statements returning different error codes is also possible
//     (We only ever consider constant value expressions. Everything else would be hard to impossible to track.)
// If the Code() method returns a single struct field:
//     Find and return the field position and identifier
//         Position needed for tracking creation with a constructor
//         Identifier needed for creation with named constructor and tracking assignments to the field
// All other return statements are marked as invalid by emitting diagnostics.
func analyseCodeMethod(pass *analysis.Pass, spec *ast.TypeSpec, funcDecl *ast.FuncDecl, receiver *ast.Ident) *ErrorType {
	constants := Set()
	var fieldName *ast.Ident
	ast.Inspect(funcDecl, func(node ast.Node) bool {
		switch node := node.(type) {
		case *ast.FuncLit:
			return false // Were not interested in return statements of nested function literals
		case *ast.ReturnStmt:
			if node.Results == nil || len(node.Results) != 1 {
				panic("should be unreachable: we already know that the method returns a single value. Return statements that don't do so should lead to a compile time error.")
			}

			returnResult := astutil.Unparen(node.Results[0])

			// If the return statement returns a constant string value:
			// Check if it is a valid error code and if so add it to the error code constants.
			returnType := pass.TypesInfo.Types[returnResult]
			if returnType.Value != nil {
				value, err := getErrorCodeFromConstant(returnType.Value)
				if err == nil {
					constants.Add(value)
				} else {
					pass.ReportRangef(node, "%v", err)
				}
				return false
			}

			// Otherwise check if a single field is returned.
			// Make sure that always the same field is returned and otherwise emit a diagnostic.
			expression, ok := returnResult.(*ast.SelectorExpr)
			if ok && receiver != nil {
				ident, ok := astutil.Unparen(expression.X).(*ast.Ident)
				if ok && ident.Obj == receiver.Obj {
					if fieldName == nil {
						fieldName = expression.Sel
					} else if fieldName.Name != expression.Sel.Name {
						pass.ReportRangef(node, "only single field allowed: cannot return field %q because field %q was returned previously", expression.Sel.Name, fieldName.Name)
					}
					return false
				}
			}

			pass.ReportRangef(node, "function %q should always return a string constant or a single field", funcDecl.Name.Name)
		}
		return true
	})

	var field *ErrorCodeField
	if fieldName != nil {
		position := getFieldPosition(spec, fieldName)
		if position >= 0 {
			field = &ErrorCodeField{fieldName.Name, position}
		} else {
			pass.Reportf(funcDecl.Pos(), "returned field %q is not a valid error code field (promoted fields are not supported currently, but might be added in the future)", fieldName)
		}
	}

	if len(constants) == 0 && field == nil {
		// In this case errors are already reported:
		// The signature of the Code() method requires at least one return statement in its implementation.
		// The return statements are all analysed and only if all are invalid this branch is entered.
		return nil
	}

	return &ErrorType{Codes: constants.Slice(), Field: field}
}

// getFieldPosition gets the position of the given field in the error struct.
func getFieldPosition(errorTypeSpec *ast.TypeSpec, fieldName *ast.Ident) int {
	errorType, ok := errorTypeSpec.Type.(*ast.StructType)
	if !ok || errorType.Fields.List == nil {
		return -1
	}

	i := 0
	for _, field := range errorType.Fields.List {
		if field.Names == nil {
			i++
			continue
		}

		for _, name := range field.Names {
			if name.Name == fieldName.Name {
				return i
			}
			i++
		}
	}

	return -1
}

// analyseMethodsOfErrorType looks at all methods of the given error type
// and makes sure there are no invalid assingments to the error code field.
func analyseMethodsOfErrorType(pass *analysis.Pass, lookup *funcLookup, errorType *ErrorType, err types.Type) {
	// Return early if there is no error code field.
	if errorType.Field == nil {
		return
	}

	assignedCodes := Set()

	errorMethods := collectMethodsForErrorType(pass, lookup, err)
	for _, method := range errorMethods {
		// Only consider methods that have a named receiver,
		// because only for those assignments to a field are possible.
		receivers := method.Recv.List[0]
		if len(receivers.Names) != 1 {
			continue
		}
		receiver := receivers.Names[0]

		ast.Inspect(method, func(node ast.Node) bool {
			assignment, ok := node.(*ast.AssignStmt)
			if !ok {
				return true
			}

			newCodes := findCodesAssignedToErrorCodeField(pass, lookup, &funcDefinition{method, nil}, errorType, receiver, assignment)
			assignedCodes = Union(assignedCodes, newCodes)

			return false
		})
	}

	// If more error codes are found, add them to the given error type.
	if len(assignedCodes) > 0 {
		codes := Union(SliceToSet(errorType.Codes), assignedCodes)
		errorType.Codes = codes.Slice()
	}
}

// collectMethodsForErrorType finds all methods defined for the given error type in the current package.
func collectMethodsForErrorType(pass *analysis.Pass, lookup *funcLookup, err types.Type) []*ast.FuncDecl {
	namedErr := getNamedType(err)
	if namedErr == nil {
		return nil
	}

	// Only consider method names that were discovered by the type checker.
	result := make([]*ast.FuncDecl, 0, namedErr.NumMethods())
	for i := 0; i < namedErr.NumMethods(); i++ {
		methodName := namedErr.Method(i).Name()
		methods, ok := lookup.methods[methodName]
		if !ok {
			continue
		}

		for _, funcDecl := range methods {
			receiverField := funcDecl.Recv.List[0]
			if !errorTypesSubset(pass.TypesInfo.TypeOf(receiverField.Type), err) {
				continue
			}
			result = append(result, funcDecl)
		}
	}

	return result
}
