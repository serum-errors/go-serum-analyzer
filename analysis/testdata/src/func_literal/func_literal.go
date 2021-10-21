package funcliteral

import "func_literal/inner"

// Errors:
//
//    - some-error --
func AssignmentInLambda() *Error { // want AssignmentInLambda:"ErrorCodes: some-error"
	var err *Error
	func() {
		err = &Error{"some-error"}
	}()
	func() {
		err := &Error{"no-error"} // shadowed!
		_ = err
	}()
	return err
}

func namedFunction() error {
	return &Error{"function-1-error"}
}

// Errors:
//
//    - lambda-1-error --
//    - lambda-2-error --
//    - lambda-3-error --
//    - lambda-4-error --
//    - lambda-5-error --
//    - lambda-6-error --
//    - lambda-7-error --
//    - function-1-error --
//    - function-2-error --
//    - other-function-error --
func LambdasReturningErrors() error { // want LambdasReturningErrors:"ErrorCodes: function-1-error function-2-error lambda-1-error lambda-2-error lambda-3-error lambda-4-error lambda-5-error lambda-6-error lambda-7-error other-function-error"
	var getError func() error = func() error {
		return &Error{"lambda-4-error"}
	}

	switch {
	case true:
		getError = func() error {
			return &Error{"lambda-2-error"}
		}
	case true:
		getError = namedFunction
	case true:
		getError = namedFunctionInOtherFile
	case true:
		getError = inner.NamedFunction
	}

	getError2 := func() error {
		return &Error{"lambda-5-error"}
	}
	_, getError3 := true, func() error {
		return &Error{"lambda-6-error"}
	}
	if false {
		getError3 = func() error {
			return &Error{"lambda-7-error"}
		}
	} else {
		getError3 = getError2
	}

	switch {
	case true:
		return getError()
	case true:
		return getError3()
	case true:
		return func() error {
			return &Error{"lambda-1-error"}
		}()
	case true:
		return func() error {
			return func() *Error {
				return &Error{"lambda-3-error"}
			}()
		}()
	}
	return nil
}

func returningLambda() func() *Error {
	return func() *Error {
		return &Error{"lambda-error"}
	}
}

// Errors: none
func OutOfBounds() *Error { // want OutOfBounds:"ErrorCodes:"
	switch {
	case true:
		return &Error{func() string { return "other-error" }()} // want "error code field has to be instantiated by constant value"
	case true:
		return func() func() *Error { // want "invalid error source: definition of the unnamed function could not be found"
			return func() *Error {
				return &Error{"lambda-error"}
			}
		}()()
	case true:
		return returningLambda()() // want "invalid error source: definition of the unnamed function could not be found"
	case true:
		var lambda func() *Error = returningLambda() // want `assignment to variable "lambda" can only be an identifier or function literal`
		return lambda()
	case true:
		funcLit := returningLambda() // want `assignment to variable "funcLit" can only be an identifier or function literal`
		return funcLit()
	case true:
		err := &Error{"context-error"}
		return func() *Error {
			return err // want "returned error may not be a parameter, global variable or other variables declared outside of the function body"
		}()
	}
	return nil
}

type Error struct { // want Error:`ErrorType{Field:{Name:"TheCode", Position:0}, Codes:}`
	TheCode string
}

func (e *Error) Code() string               { return e.TheCode }
func (e *Error) Message() string            { return e.TheCode }
func (e *Error) Details() map[string]string { return nil }
func (e *Error) Cause() error               { return nil }
func (e *Error) Error() string              { return e.Message() }
