package inner2

import "interfaces/inner1"

type ImplementOuter2 struct{}

// Errors:
//
//    - interface-1-error --
func (ImplementOuter2) InterfaceMethod1() error {
	return &Error{"interface-1-error"}
}

// Errors:
//
//    - interface-3-error --
func (ImplementOuter2) InterfaceMethod2(a, b string) error {
	return &Error{"interface-3-error"}
}

func (ImplementOuter2) MethodWithoutError(a, b string) string {
	return a
}

type ImplementOuterPointer2 struct{}

// Errors:
//
//    - interface-2-error --
func (*ImplementOuterPointer2) InterfaceMethod1() error {
	return &Error{"interface-2-error"}
}

// Errors:
//
//    - interface-4-error --
func (*ImplementOuterPointer2) InterfaceMethod2(a, b string) error {
	return &Error{"interface-4-error"}
}

func (*ImplementOuterPointer2) MethodWithoutError(a, b string) string {
	return a
}

type ImplementInner1Interface1 struct{}

// Errors:
//
//    - interface-2-error --
func (ImplementInner1Interface1) Inner1Method1() error { // want Inner1Method1:"ErrorCodes: interface-2-error"
	return &Error{"interface-2-error"}
}

// Errors: none
func (ImplementInner1Interface1) Inner1Method2(a, b string) *inner1.Error { // want Inner1Method2:"ErrorCodes:"
	return nil
}

func (ImplementInner1Interface1) Inner1MethodWithoutError(a, b string) string {
	return a + b
}

type Error struct {
	TheCode string
}

func (e *Error) Code() string               { return e.TheCode }
func (e *Error) Message() string            { return e.TheCode }
func (e *Error) Details() map[string]string { return nil }
func (e *Error) Cause() error               { return nil }
func (e *Error) Error() string              { return e.Message() }
