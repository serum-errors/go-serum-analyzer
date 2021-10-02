package main

func ErrorNotLast() (error, int) { // want "error should be returned as the last argument"
	return &Error{"hello-error"}, 0
}

// CallToInvalidFunction calls a function which has an error as non-last return argument.
//
// Errors:
//
//    - zonk-error -- is always returned
func CallToInvalidFunction() error { // want CallToInvalidFunction:"ErrorCodes: zonk-error"
	e, _ := ErrorNotLast() // want "unsupported: tracking error codes for function call with error as non-last return argument"
	if false {
		return e
	}
	return &Error{"zonk-error"}
}

// ReturnInvalidError returns an error that does not define a Code() method.
//
// Errors:
//
//    - hello-error -- might be returned by this function
func ReturnInvalidError() error { // want ReturnInvalidError:"ErrorCodes: hello-error"
	if false {
		return &Error{"hello-error"}
	}
	return &InvalidError{} // want "expression does not define an error code"
}

type InvalidError struct{}

func (e *InvalidError) Error() string { return "InvalidError" }

// InvalidErrorCodeFormat returns an error with invalid error code.
//
// Errors:
//
//    - hello-error -- might be returned
func InvalidErrorCodeFormat() error { // want InvalidErrorCodeFormat:"ErrorCodes: hello-error"
	return invalidErrorCodeFormat()
}

func invalidErrorCodeFormat() error {
	switch {
	case true:
		return &Error{"5-invalid-error"} // want "error code has invalid format: should match .*"
	case true:
		return &Error{"-invalid-error"} // want "error code has invalid format: should match .*"
	case true:
		return &Error{"invalid-error-"} // want "error code has invalid format: should match .*"
	case true:
		return &Error{"invalid-(chars)-error"} // want "error code has invalid format: should match .*"
	case true:
		return &Error{"invalid error"} // want "error code has invalid format: should match .*"
	default:
		return &Error{"hello-error"} // valid
	}
}

// Errors:
//
//    - some-error --
func InvalidErrorFromParameter(p error) error { // want InvalidErrorFromParameter:"ErrorCodes: some-error"
	switch {
	case true:
		return p // want "returned error may not be a parameter, receiver or global variable"
	case true:
		x := p // want "returned error may not be a parameter, receiver or global variable"
		return x
	}
	return &Error{"some-error"}
}

// Errors:
//
//    - some-error --
func (e *Error) InvalidErrorFromReceiver() error { // want InvalidErrorFromReceiver:"ErrorCodes: some-error"
	switch {
	case true:
		return e // want "returned error may not be a parameter, receiver or global variable"
	case true:
		x := e // want "returned error may not be a parameter, receiver or global variable"
		return x
	}
	return &Error{"some-error"}
}

var globalError = &Error{"global-error"}

// Errors:
//
//    - some-error --
func InvalidErrorFromGlobal() error { // want InvalidErrorFromGlobal:"ErrorCodes: some-error"
	switch {
	case true:
		return globalError // want "returned error may not be a parameter, receiver or global variable"
	case true:
		x := globalError // want "returned error may not be a parameter, receiver or global variable"
		return x
	}
	return &Error{"some-error"}
}

func getLambda() func() error {
	return func() error {
		return &Error{"lambda-error"}
	}
}

// Errors:
//
//    - some-error --
func InvalidErrorFromLambda() error { // want InvalidErrorFromLambda:"ErrorCodes: some-error"
	if false {
		return getLambda()() // want "unnamed functions are not yet supported in error code analysis"
	}
	return &Error{"some-error"}
}
