package typeassertion

import (
	"math/rand"
)

type Receiver struct{}

// Errors:
//
//    - interface-error -- always
func (Receiver) Method() error { // want Method:"ErrorCodes: interface-error"
	return &Error{"interface-error"}
}

// Errors:
//
//    - standard-error -- always
func StandardError() error { // want StandardError:"ErrorCodes: standard-error"
	return &Error{"standard-error"}
}

// EmptyInterfaceError isn't detected by serum-analyzer as an error return type
func EmptyInterfaceError() interface{} {
	return &Error{"empty-interface-error"}
}

// Errors:
//
//   - maybe-error -- sometimes
func MaybeError() error { // want MaybeError:"ErrorCodes: maybe-error"
	if rand.Float64() < 0.5 {
		return &Error{"maybe-error"}
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

type ErrorInterface interface {
	error
	Code() string
}
