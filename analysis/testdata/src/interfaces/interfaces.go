package interfaces

import (
	"interfaces/inner1"
	"interfaces/inner2"
)

type SomeInterface interface {
	// InterfaceMethod1 is a method returning an error with error codes declared in doc.
	//
	// Errors:
	//
	//    - interface-1-error -- could potentially be returned
	//    - interface-2-error --
	InterfaceMethod1() error // want InterfaceMethod1:"ErrorCodes: interface-1-error interface-2-error"

	// Errors:
	//
	//    - interface-3-error -- could potentially be returned
	//    - interface-4-error --
	InterfaceMethod2(a, b string) error // want InterfaceMethod2:"ErrorCodes: interface-3-error interface-4-error"

	// MethodWithoutError is just some method that does not return an error,
	// to test if it is correctly ignored in the analysis.
	MethodWithoutError(a, b string) string
}

// Errors:
//
//    - interface-1-error --
//    - interface-2-error --
//    - interface-3-error --
//    - interface-4-error --
func InterfaceParam(some SomeInterface, a, b string) error { // want InterfaceParam:"ErrorCodes: interface-1-error interface-2-error interface-3-error interface-4-error"
	if false {
		return some.InterfaceMethod1()
	}
	return some.InterfaceMethod2(a, b)
}

// Errors:
//
//    - interface-1-error --
//    - interface-2-error --
//    - interface-3-error --
//    - interface-4-error --
func OuterFunction(a, b string) error { // want OuterFunction:"ErrorCodes: interface-1-error interface-2-error interface-3-error interface-4-error"
	if false {
		return InterfaceParam(inner1.ImplementOuter1{}, a, b)
	}
	return InterfaceParam(inner2.ImplementOuter2{}, a, b)
}

type (
	ImplementInner1Interface1 struct{}
	ImplementInner1Interface2 struct{}
	ImplementInner1Interface3 struct{}
)

// Errors:
//
//    - interface-1-error --
func (ImplementInner1Interface1) Inner1Method1() error { // want Inner1Method1:"ErrorCodes: interface-1-error"
	return &Error{"interface-1-error"}
}

// Errors:
//
//    - interface-3-error --
func (ImplementInner1Interface1) Inner1Method2(a, b string) *inner1.Error { // want Inner1Method2:"ErrorCodes: interface-3-error"
	return &inner1.Error{"interface-3-error"}
}

func (ImplementInner1Interface1) Inner1MethodWithoutError(a, b string) string {
	return a
}

// Errors:
//
//    - interface-1-error --
func (ImplementInner1Interface2) Inner1CodeNotDeclared() error { /*
		want
			Inner1CodeNotDeclared:"ErrorCodes: interface-1-error"
			`function "Inner1CodeNotDeclared" declares error codes that are not declared in the interface "Inner1Interface2": \[interface-1-error]` */
	return &Error{"interface-1-error"}
}

func (ImplementInner1Interface3) Inner1NoCodes() error { // want `function "Inner1NoCodes" is exported, but does not declare any error codes`
	return nil
}

// Errors:
//
//    - interface-1-error -- could potentially be returned
//    - interface-2-error --
func (ImplementInner1Interface3) Inner1YesCodes() error { // want Inner1YesCodes:"ErrorCodes: interface-1-error interface-2-error"
	if false {
		return &Error{"interface-1-error"}
	}
	return &Error{"interface-2-error"}
}

type Error struct { // want Error:`ErrorType{Field:{Name:"TheCode", Position:0}, Codes:}`
	TheCode string
}

func (e *Error) Code() string               { return e.TheCode }
func (e *Error) Message() string            { return e.TheCode }
func (e *Error) Details() map[string]string { return nil }
func (e *Error) Cause() error               { return nil }
func (e *Error) Error() string              { return e.Message() }
