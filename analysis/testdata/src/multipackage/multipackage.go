package multipackage

import (
	"multipackage/inner1"
	"multipackage/inner2"
)

// RunPackage1 is a demo function.
//
// Errors:
//
//    - hello-error -- is always returned
func RunPackage1() error { // want RunPackage1:"ErrorCodes: hello-error"
	return inner1.ExportedFunc1()
}

// RunPackage2 is a demo function
//
// Errors:
//
//    - hello-error -- is always returned
func RunPackage2() error { // want RunPackage2:"ErrorCodes: hello-error"
	return inner2.ExportedFunc2()
}

// Trap1 is a demo function.
//
// Errors:
//
//    - hello-error -- is always returned
func Trap1() error { // want Trap1:"ErrorCodes: hello-error"
	trap := returnTrap()
	return trap.returnError()
}

// Trap2 is a demo function.
//
// Errors:
//
//    - hello-error -- is always returned
func Trap2() error { // want Trap2:"ErrorCodes: hello-error"
	return returnTrap().returnError()
}

type TrapType struct{}

func (TrapType) returnError() error {
	return &Error{"hello-error"}
}

func returnTrap() TrapType {
	return TrapType{}
}

// Inner1Error returns an error from another package.
//
// Errors:
//
//    - inner1-error -- is always returned
func Inner1Error() error { // want Inner1Error:"ErrorCodes: inner1-error"
	return &inner1.Error{"inner1-error"}
}

// Inner2Error returns an error from another package.
//
// Errors:
//
//    - inner2-error -- is always returned
func Inner2Error() error { // want Inner2Error:"ErrorCodes: inner2-error"
	return &inner2.Error{"inner2-error"}
}

// Inner1UnusedError returns an error from another package,
// which is not returned by any function in the inner package.
//
// Errors:
//
//    - inner1-unused-error -- is always returned
func Inner1UnusedError() error { // want Inner1UnusedError:"ErrorCodes: inner1-unused-error"
	return &inner1.UnusedError{"inner1-unused-error"}
}

// Inner2UnusedError returns an error from another package,
// which is not returned by any function in the inner package.
//
// Errors:
//
//    - inner2-unused-error -- is always returned
func Inner2UnusedError() error { // want Inner2UnusedError:"ErrorCodes: inner2-unused-error"
	return &inner2.UnusedError{"inner2-unused-error"}
}

// Errors:
//
//    - x-error --
func CallToUndeclared1() error { // want CallToUndeclared1:"ErrorCodes: x-error"
	if true {
		return inner1.CodeNotDeclared() // want `function "CodeNotDeclared" in package "inner1" does not declare error codes`
	}
	return &Error{"x-error"}
}

// Errors:
//
//    - x-error --
func CallToUndeclared2() error { // want CallToUndeclared2:"ErrorCodes: x-error"
	if true {
		return inner2.CodeNotDeclared() // want `function "CodeNotDeclared" in package "inner2" does not declare error codes`
	}
	return &Error{"x-error"}
}

// Errors:
//
//    - x-error --
func CallToUndeclared3() error { // want CallToUndeclared3:"ErrorCodes: x-error"
	if true {
		object := inner1.SomeType{}
		return object.CodeNotDeclared() // want "called function does not declare error codes"
	}
	return &Error{"x-error"}
}

// Errors:
//
//    - x-error --
func CallToUndeclared4() error { // want CallToUndeclared4:"ErrorCodes: x-error"
	if true {
		object := inner2.SomeType{}
		return object.CodeNotDeclared() // want "called function does not declare error codes"
	}
	return &Error{"x-error"}
}

// Errors:
//
//    - string-error --
func DoString() error { // want DoString:"ErrorCodes: string-error"
	return inner1.StringError("some message about the error")
}

type Error struct { // want Error:`ErrorType{Field:{Name:"TheCode", Position:0}, Codes:}`
	TheCode string
}

func (e *Error) Code() string               { return e.TheCode }
func (e *Error) Message() string            { return e.TheCode }
func (e *Error) Details() map[string]string { return nil }
func (e *Error) Cause() error               { return nil }
func (e *Error) Error() string              { return e.Message() }
