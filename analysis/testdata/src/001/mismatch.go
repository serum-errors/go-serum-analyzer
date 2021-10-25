package main

// UnusedCode declares an error which is never returned.
//
// Errors:
//
//    - unused-error -- never returned
func UnusedCode() error { /*
		want
			UnusedCode:"ErrorCodes: unused-error"
			`function "UnusedCode" has a mismatch of declared and actual error codes: unused codes: \[unused-error]` */
	return nil
}

// MissingCode returns an error which is not declared.
//
// Errors:
//
//    - hello-error -- error that might be returned
func MissingCode() error { /*
		want
			MissingCode:"ErrorCodes: hello-error"
			`function "MissingCode" has a mismatch of declared and actual error codes: missing codes: \[missing-error]` */
	if false {
		return &Error{"hello-error"}
	}
	return &Error{"missing-error"}
}

// MissingAndUnusedCode returns an error which is not declared,
// and declares an error which is never returned.
//
// Errors:
//
//    - unused-error -- never returned
func MissingAndUnusedCode() error { /*
		want
			MissingAndUnusedCode:"ErrorCodes: unused-error"
			`function "MissingAndUnusedCode" has a mismatch of declared and actual error codes: missing codes: \[missing-error] unused codes: \[unused-error]` */
	return &Error{"missing-error"}
}

// MultipleMismatchedCodes has multiple error code mismatches.
//
// Errors:
//
//    - hello-error     -- error that might be returned
//    - unused-error-ab -- never returned
//    - unused-error-aa -- never returned
func MultipleMismatchedCodes() error { /*
		want
			MultipleMismatchedCodes:"ErrorCodes: hello-error unused-error-aa unused-error-ab"
			`function "MultipleMismatchedCodes" has a mismatch of declared and actual error codes: missing codes: \[missing-error-a missing-error-ab missing-error-cc] unused codes: \[unused-error-aa unused-error-ab]` */
	switch {
	case true:
		return &Error{"missing-error-ab"}
	case true:
		return &Error{"missing-error-a"}
	case true:
		return &Error{"missing-error-cc"}
	default:
		return &Error{"hello-error"}
	}
}

// UnusedParam declares an error code parameter which is never used.
//
// Errors:
//
//    - param: code -- unused param
func UnusedParam(code string) error { /*
		want
			UnusedParam:"ErrorConstructor: {CodeParamPosition:0}"
			UnusedParam:"ErrorCodes:"
			`function "UnusedParam" has a mismatch of declared and actual error codes: unused param: "code"` */
	_ = code
	return nil
}

// MissingParam returns an error using the parameter "code" as the error code,
// but never declares an error code parameter
//
// Errors: none
func MissingParam(code string) error { /*
		want
			MissingParam:"ErrorCodes:"
			`function "MissingParam" has a mismatch of declared and actual error codes: missing param: "code"` */
	return &Error{code}
}

// WrongParam declares the wrong parameter as error code parameter.
//
// Errors:
//
//    - param: code2 -- unused param
func WrongParam(code1, code2 string) error { /*
		want
			WrongParam:"ErrorConstructor: {CodeParamPosition:1}"
			WrongParam:"ErrorCodes:  Param:{Position:1}"
			`function "WrongParam" has a mismatch of declared and actual error codes: missing param: "code1" unused param: "code2"` */
	_ = code2
	return &Error{code1}
}

// MismatchedCodesAndParams has multiple error code and error code parameter mismatches.
//
// Errors:
//
//    - hello-error       -- error that might be returned
//    - unused-error      -- never returned
//    - param: codeUnused -- never used
func MismatchedCodesAndParams(codeUnused, codeMissing string) error { /*
		want
			MismatchedCodesAndParams:"ErrorConstructor: {CodeParamPosition:0}"
			MismatchedCodesAndParams:"ErrorCodes: hello-error unused-error"
			`function "MismatchedCodesAndParams" has a mismatch of declared and actual error codes: missing codes: \[missing-1-error missing-2-error] unused codes: \[unused-error] missing param: "codeMissing" unused param: "unusedCode"` */
	switch {
	case true:
		return &Error{"missing-1-error"}
	case true:
		return &Error{"missing-2-error"}
	case true:
		return &Error{codeMissing}
	default:
		return &Error{"hello-error"}
	}
}
