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
//    - function-1-error --
//    - function-2-error --
func LambdasReturningErrors() error { // want LambdasReturningErrors:"ErrorCodes: function-1-error function-2-error lambda-1-error lambda-2-error"
	var getError func() error
	switch {
	case true:
		getError = func() error {
			return &Error{"lambda-2-error"}
		}
	case true:
		getError = namedFunction
	case true:
		getError = inner.NamedFunction
	}

	switch {
	case true:
		return getError()
	case true:
		return func() error {
			return &Error{"lambda-1-error"}
		}()
	}
	return nil
}

// Errors: none
func OutOfBounds() *Error { // want OutOfBounds:"ErrorCodes:"
	switch {
	case true:
		return &Error{func() string { return "other-error" }()} // want "error code field has to be instantiated by constant value"
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
