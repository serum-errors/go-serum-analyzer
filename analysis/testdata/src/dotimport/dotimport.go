package dotimport

import (
	. "dotimport/inner1"
	. "dotimport/inner2"
)

// RunPackage1 is a demo function.
//
// Errors:
//
//    - hello-error -- is always returned
func RunPackage1() error { // want RunPackage1:"ErrorCodes: hello-error"
	return ExportedFunc1()
}

// RunPackage2 is a demo function
//
// Errors:
//
//    - hello-error -- is always returned
func RunPackage2() error { // want RunPackage2:"ErrorCodes: hello-error"
	return ExportedFunc2()
}

// ReturnInner1Error returns an error from another package.
//
// Errors:
//
//    - inner1-error -- is always returned
func ReturnInner1Error() error { // want ReturnInner1Error:"ErrorCodes: inner1-error"
	return &Inner1Error{"inner1-error"}
}

// ReturnInner2Error returns an error from another package.
//
// Errors:
//
//    - inner2-error -- is always returned
func ReturnInner2Error() error { // want ReturnInner2Error:"ErrorCodes: inner2-error"
	return &Inner2Error{"inner2-error"}
}

// ReturnInner1UnusedError returns an error from another package,
// which is not returned by any function in the inner package.
//
// Errors:
//
//    - inner1-unused-error -- is always returned
func ReturnInner1UnusedError() error { // want ReturnInner1UnusedError:"ErrorCodes: inner1-unused-error"
	return &Inner1UnusedError{"inner1-unused-error"}
}

// ReturnInner2UnusedError returns an error from another package,
// which is not returned by any function in the inner package.
//
// Errors:
//
//    - inner2-unused-error -- is always returned
func ReturnInner2UnusedError() error { // want ReturnInner2UnusedError:"ErrorCodes: inner2-unused-error"
	return &Inner2UnusedError{"inner2-unused-error"}
}

// Errors:
//
//    - x-error --
func CallToUndeclared1() error { // want CallToUndeclared1:"ErrorCodes: x-error"
	if true {
		return CodeNotDeclared1() // want `function "CodeNotDeclared1" in dot-imported package does not declare error codes`
	}
	return &Error{"x-error"}
}

// Errors:
//
//    - x-error --
func CallToUndeclared2() error { // want CallToUndeclared2:"ErrorCodes: x-error"
	if true {
		return CodeNotDeclared2() // want `function "CodeNotDeclared2" in dot-imported package does not declare error codes`
	}
	return &Error{"x-error"}
}

// Errors:
//
//    - x-error --
func CallToUndeclared3() error { // want CallToUndeclared3:"ErrorCodes: x-error"
	if true {
		object := SomeType1{}
		return object.CodeNotDeclared() // want "called function does not declare error codes"
	}
	return &Error{"x-error"}
}

// Errors:
//
//    - x-error --
func CallToUndeclared4() error { // want CallToUndeclared4:"ErrorCodes: x-error"
	if true {
		object := SomeType2{}
		return object.CodeNotDeclared() // want "called function does not declare error codes"
	}
	return &Error{"x-error"}
}

type Error struct { // want Error:`ErrorType{Field:{Name:"TheCode", Position:0}, Codes:}`
	TheCode string
}

func (e *Error) Code() string               { return e.TheCode }
func (e *Error) Message() string            { return e.TheCode }
func (e *Error) Details() map[string]string { return nil }
func (e *Error) Cause() error               { return nil }
func (e *Error) Error() string              { return e.Message() }
