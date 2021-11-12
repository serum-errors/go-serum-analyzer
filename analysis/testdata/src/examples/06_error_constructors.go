package examples

// NewError creates an Error with the given error code.
//
// Errors:
//
//    - param: code -- error code parameter
func NewError(code string) *Error { // want NewError:"ErrorConstructor: {CodeParamPosition:0}" NewError:"ErrorCodes:"
	return &Error{code}
}

// NewError3 creates an Error3 with the given error code.
//
// Errors:
//
//    - param: c               -- error code parameter, which is set if flag is true
//    - examples-error-unknown --
//    - examples-error-flagged --
func NewError3(flag bool, c string) *Error3 { // want NewError3:"ErrorConstructor: {CodeParamPosition:1}" NewError3:"ErrorCodes: examples-error-flagged examples-error-unknown"
	err := &Error3{flag, "examples-error-unknown"}
	if flag {
		err.code = c
	}
	return err
}

// Errors:
//
//    - examples-error-not-implemented --
func CallConstructor() error { // want CallConstructor:"ErrorCodes: examples-error-not-implemented"
	return NewError("examples-error-not-implemented")
}

// Errors:
//
//    - param: errorCode       --
//    - examples-error-unknown --
//    - examples-error-flagged --
func NewGeneralError(flag bool, errorCode string) error { // want NewGeneralError:"ErrorConstructor: {CodeParamPosition:1}" NewGeneralError:"ErrorCodes: examples-error-flagged examples-error-unknown"
	if flag {
		return NewError3(flag, errorCode)
	}
	return NewError(errorCode)
}
