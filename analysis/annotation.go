package analysis

import (
	"fmt"
	"go/ast"
	"strings"
)

const (
	annotationIndicatorReturnStmt = "Error Codes"
)

type (
	annotationReturnStmt struct {
		shouldOverwrite bool
		overwrite       CodeSet
		addCodes        CodeSet
		subCodes        CodeSet
	}
)

// getReturnStmtAnnotations finds and returns annotations for the given return statement.
//
// An annotation:
//   - is a comment directly above the return statement
//   - starts with "Error Codes"
//   - is unique (only one annotation per return statement)
//
// The annotation changes what error codes are reported for the targeted return statement:
//   - "Error Codes  = code1, code2, ..." overwrites any returned codes to be the codes defined after the '='.
//   - "Error Codes += code1, code2, ..." adds the defined codes to the returned codes.
//   - "Error Codes -= code1, code2, ..." subtracts the defined codes from the returned codes.
//   - "Error Codes -code1, +code2, +code3, ..." adds and/or substracts the defined codes from the returned codes,
//     depending on if they have a '+' or a '-' prefix.
func getReturnStmtAnnotations(c *context, stmt *ast.ReturnStmt) *annotationReturnStmt {
	pass := c.pass

	comments, ok := c.comments[stmt]
	if !ok || len(comments) == 0 {
		return nil
	}

	var result *annotationReturnStmt

	comment := comments[0].Text()
	for _, line := range strings.Split(comment, "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, annotationIndicatorReturnStmt) {
			continue
		}
		line = strings.TrimPrefix(line, annotationIndicatorReturnStmt)

		var err error
		line = strings.TrimSpace(line)
		if len(line) == 0 {
			pass.ReportRangef(stmt, "error in annotation: expected '=', '+=', '-=', '+code', or '-code' after '%s' indicator", annotationIndicatorReturnStmt)
			return nil
		}

		if result != nil {
			pass.ReportRangef(stmt, "found multiple annotations for the same return statement: only one is allowed")
		}
		result = &annotationReturnStmt{false, Set(), Set(), Set()}

		switch line[0] {
		case '=':
			result.overwrite, err = extractCodesFromStringAnnotation(line[1:])
			result.shouldOverwrite = true
		case '+':
			if len(line) >= 2 && line[1] == '=' {
				result.addCodes, err = extractCodesFromStringAnnotation(line[2:])
			} else {
				err = extractCodeModificationsFromStringAnnotation(result, line[1:])
			}
		case '-':
			if len(line) >= 2 && line[1] == '=' {
				result.subCodes, err = extractCodesFromStringAnnotation(line[2:])
			} else {
				err = extractCodeModificationsFromStringAnnotation(result, line[1:])
			}
		default:
			err = fmt.Errorf("error in annotation: expected '=', '+=', '-=', '+code', or '-code' after '%s' indicator", annotationIndicatorReturnStmt)
		}

		if err != nil {
			pass.ReportRangef(stmt, "%v", err)
			return nil
		}
	}

	return result
}

// extractCodesFromStringAnnotation creates an error code set from a comma separated list of error codes.
func extractCodesFromStringAnnotation(annotation string) (CodeSet, error) {
	result := Set()

	for _, code := range strings.Split(annotation, ",") {
		code = strings.TrimSpace(code)
		if err := checkErrorCodeValid(code); err != nil {
			return nil, fmt.Errorf("invalid error code in annotation: %v", err)
		}

		result.Add(code)
	}

	return result, nil
}

// extractCodeModificationsFromStringAnnotation finds codes that should be added or removed from a comma separated list.
//
// Codes that should be added have a '+' prefix, while codes that should be removed have a '-' prefix.
func extractCodeModificationsFromStringAnnotation(result *annotationReturnStmt, annotation string) error {
	for _, code := range strings.Split(annotation, ",") {
		code = strings.TrimSpace(code)
		if len(code) == 0 || (code[0] != '+' && code[0] != '-') {
			return fmt.Errorf("invalid error code in annotation: code has to start with '+' or '-'")
		}

		if err := checkErrorCodeValid(code[1:]); err != nil {
			return fmt.Errorf("invalid error code in annotation: %v", err)
		}

		switch code[0] {
		case '+':
			result.addCodes.Add(code[1:])
		case '-':
			result.subCodes.Add(code[1:])
		}
	}

	return nil
}
