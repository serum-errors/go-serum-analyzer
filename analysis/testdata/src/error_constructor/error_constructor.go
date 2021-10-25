package errorconstructor

// NewError is a constructor for Error using the code parameter.
//
// Errors:
//
//    - param: code   -- is used if the provided flag is true
//    - unknown-error -- is used otherwise
func NewError(code string, flag bool) error { // want NewError:"ErrorConstructor: {CodeParamPosition:0}" NewError:"ErrorCodes: unknown-error"
	if flag {
		return &Error{code}
	}
	return &Error{"unknown-error"}
}

// NewError is a constructor for Error using the code parameter.
//
// Errors:
//
//    - param: code --
func NewError2(code string) error { // want NewError:"ErrorConstructor: {CodeParamPosition:0}" NewError:"ErrorCodes:"
	return &Error{code}
}

// MissingParam returns an error using the parameter "code" as the error code,
// but never declares an error code parameter
//
// Errors: none
func MissingParam(code string) error { // want MissingParam:"ErrorCodes:"
	return &Error{code} // want `require an error code parameter declaration to use "code" as an error code`
}

// WrongParam declares the wrong parameter as error code parameter.
//
// Errors:
//
//    - param: code2 -- unused param
func WrongParam(code1, code2 string) error { /*
		want
			WrongParam:"ErrorConstructor: {CodeParamPosition:1}"
			WrongParam:"ErrorCodes:" */
	_ = code2
	return &Error{code1} // want `require an error code parameter declaration to use "code1" as an error code`
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
			`function "MismatchedCodesAndParams" has a mismatch of declared and actual error codes: missing codes: \[missing-1-error missing-2-error] unused codes: \[unused-error]` */
	switch {
	case true:
		return &Error{"missing-1-error"}
	case true:
		return &Error{"missing-2-error"}
	case true:
		return &Error{codeMissing} // want `require an error code parameter declaration to use "codeMissing" as an error code`
	default:
		return &Error{"hello-error"}
	}
}

// Errors:
//
//    - param: code --
//    - some-error  --
func CorrectFieldAssign(code string, flag bool) error { // want InvalidFieldAssign:"ErrorConstructor: {ErrorCodeParam:0}" InvalidFieldAssign:"ErrorCodes: some-error"
	err := &Error{"some-error"}
	if flag {
		err.TheCode = code
	}
	return err
}

// Errors:
//
//    - some-error  --
func InvalidFieldAssign(code string, flag bool) error { // want InvalidFieldAssign:"ErrorCodes: some-error"
	err := &Error{"some-error"}
	if flag {
		err.TheCode = code // want `require an error code parameter declaration to use "code" as an error code`
	}
	return err
}

type Error struct { // want Error:`ErrorType{Field:{Name:"TheCode", Position:0}, Codes:}`
	TheCode string
}

func (e *Error) Code() string               { return e.TheCode }
func (e *Error) Message() string            { return e.TheCode }
func (e *Error) Details() map[string]string { return nil }
func (e *Error) Cause() error               { return nil }
func (e *Error) Error() string              { return e.Message() }
