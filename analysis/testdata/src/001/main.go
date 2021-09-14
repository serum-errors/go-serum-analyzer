package main

import (
	"fmt"
)

func main() {
	One()
}

/*
	One is a demo function.
	It calls another function that returns an error.

	Errors:

		- hello-error -- always returned.
*/
func One() error { // want One:"ErrorCodes: hello-error"
	return Two()
}

// Two is a demo function.
// It returns an error value it creates itself.
//
// Errors:
//
//    - hello-error -- is always returned.
func Two() error { // want Two:"ErrorCodes: hello-error"
	return &Error{"hello-error"}
}

func Three() *Error { // want `function "Three" is exported, but does not declare any error codes`
	return &Error{"hello-error-literal"}
}

// Four is a demo function with multiple returns.
//
// Errors:
//
//    - zonk-error -- is always returned.
func Four() (string, error) { // want Four:"ErrorCodes: zonk-error"
	return "something", &Error{"zonk-error"}
}

// Five is a demo function with multiple returns
// which is returning the result of another function
// (which *also* has multiple returns).
//
// Errors:
//
//    - zonk-error -- is always returned.
func Five() (interface{}, error) { // want Five:"ErrorCodes: zonk-error"
	return Four()
}

// Five is a demo function with multiple returns
// which is returning the result of another function
// (which *also* has multiple returns).
//
// Errors:
//
//    - hello-error -- is sometimes returned.
//    - zonk-error -- is returned at other times.
func Six(flip bool) error { // want Six:"ErrorCodes: hello-error zonk-error"
	err := Two()
	if flip {
		_, err = Four()
	}
	return err
}

// Seven is ... probably out of scope.
//
// Errors:
//
//    - hello-error -- is always returned.
func Seven() error {
	uff := &Error{}
	uff.TheCode = "hello-error"
	var err error
	err = uff
	return err
}

// Eight isn't going to fulfill what it says it will do.
// It also makes a call to another package.
//
// Errors:
//
//    - hello-error -- is a lie, won't actually happen.
func Eight() error {
	return fmt.Errorf("not a nice structural error")
}

type Error struct { // want Error:`ErrorType{Field:{Name:"TheCode", Position:0}, Codes:}`
	TheCode string
}

func (e *Error) Code() string               { return e.TheCode }
func (e *Error) Message() string            { return e.TheCode }
func (e *Error) Details() map[string]string { return nil }
func (e *Error) Cause() error               { return nil }
func (e *Error) Error() string              { return e.Message() }
