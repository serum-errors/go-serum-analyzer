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
func NewError2(code string) error { // want NewError2:"ErrorConstructor: {CodeParamPosition:0}" NewError2:"ErrorCodes:"
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
func CorrectFieldAssign(code string, flag bool) error { // want CorrectFieldAssign:"ErrorConstructor: {CodeParamPosition:0}" CorrectFieldAssign:"ErrorCodes: some-error"
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

// Errors:
//
//    - param-error   --
//    - unknown-error --
func CallConstructor() error { // want CallConstructor:"ErrorCodes: param-error unknown-error"
	return NewError("param-error", false)
}

// Errors:
//
//    - param: code   --
//    - unknown-error --
func CallConstructorFromConstructor(code string) error { // want CallConstructorFromConstructor:"ErrorConstructor: {CodeParamPosition:0}" CallConstructorFromConstructor:"ErrorCodes: unknown-error"
	return NewError(code, false)
}

// Errors:
//
//    - param: code --
//    - param-error --
func RecursiveConstructor(code string) error { // want RecursiveConstructor:"ErrorConstructor: {CodeParamPosition:0}" RecursiveConstructor:"ErrorCodes: param-error"
	if false {
		return RecursiveConstructor(code)
	}
	return NewError2("param-error")
}

// Errors: none
func InvalidCallConstructor() error { // want InvalidCallConstructor:"ErrorCodes:"
	var someCode string = "some-error"
	return NewError2(someCode) // want `error code has to be constant value or error code parameter`
}

// Errors: none
func InvalidCallConstructor2(someCode string) error { // want InvalidCallConstructor2:"ErrorCodes:"
	return NewError2(someCode) // want `require an error code parameter declaration to use "someCode" as an error code`
}

// Errors: none
func InvalidCallConstructor3() error { // want InvalidCallConstructor3:"ErrorCodes:"
	var postFix string = "-error"
	return NewError2("another-" + postFix) // want `error code has to be constant value or error code parameter`
}

// Errors:
//
//    - param: code --
func InvalidUseOfCodeParam(code string) error { // want InvalidUseOfCodeParam:"ErrorConstructor: {CodeParamPosition:0}" InvalidUseOfCodeParam:"ErrorCodes:"
	return func() error {
		return &Error{code} // want "error code has to be constant value or error code parameter"
	}()
}

// Errors:
//
//    - param: code --
func InvalidUseOfCodeParam2(code string) error { // want InvalidUseOfCodeParam2:"ErrorConstructor: {CodeParamPosition:0}" InvalidUseOfCodeParam2:"ErrorCodes:"
	lambda := NewError2 // want `unsupported use of error constructor "NewError2"`
	return lambda(code)
}

// Errors: none
func EmptyStringConstruct() error { // want EmptyStringConstruct:"ErrorCodes:"
	return NewError2("")
}

// Errors:
//
//    - param: code -- error code parameter
//    - some-error  -- assigned to error code parameter
//    - other-error  -- indirectly assigned to error code parameter
func AssignToParam(_ int, other, code string) error { // want AssignToParam:"ErrorConstructor: {CodeParamPosition:2}" AssignToParam:"ErrorCodes: other-error some-error"
	var otherError string
	switch {
	case true:
		code = "" // allowed
	case true:
		code = "some-error" //allowed
	case true:
		code = other // want "error code parameter may not be assigned an other parameter, receiver or global variable"
	case true:
		code = "-invalid" // want "error code has invalid format: should match (.*)"
	case true:
		const constant = "other-error"
		otherError = constant
	case true:
		code = otherError
	case true:
		otherError = "-invalid" // want "error code has invalid format: should match (.*)"
	}
	return NewError2(code)
}

type ConstructorInterface interface {
	// Errors:
	//
	//    - param: code --
	NewError(code string) *Error // want "declaration of error constructors in interfaces is currently not supported"
}

type Error struct { // want Error:`ErrorType{Field:{Name:"TheCode", Position:0}, Codes:}`
	TheCode string
}

func (e *Error) Code() string               { return e.TheCode }
func (e *Error) Message() string            { return e.TheCode }
func (e *Error) Details() map[string]string { return nil }
func (e *Error) Cause() error               { return nil }
func (e *Error) Error() string              { return e.Message() }
